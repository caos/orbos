package ssh

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/caos/infrop/internal/core/logging"
	"github.com/caos/infrop/internal/core/operator"
	"github.com/caos/infrop/internal/kinds/clusters/core/infra"
	"github.com/pkg/errors"
	sshlib "golang.org/x/crypto/ssh"
)

type ProvidedCompute interface {
	ID() string
	InternalIP() (*string, error)
	ExternalIP() (*string, error)
	Remove() error
}

type compute struct {
	logger     logging.Logger
	compute    ProvidedCompute
	remoteUser string
	sshCfg     *sshlib.ClientConfig
}

func NewCompute(logger logging.Logger, comp ProvidedCompute, remoteUser string) infra.Compute {
	return &compute{
		remoteUser: remoteUser,
		logger: logger.WithFields(map[string]interface{}{
			"compute": comp.ID(),
		}),
		compute: comp,
	}
}

func (c *compute) InternalIP() (*string, error) {
	return c.compute.InternalIP()
}

func (c *compute) ExternalIP() (*string, error) {
	return c.compute.ExternalIP()
}

func (c *compute) ID() string {
	return c.compute.ID()
}

func (c *compute) Remove() error {
	return c.compute.Remove()
}

func (c *compute) Execute(env map[string]string, stdin io.Reader, cmd string) (stdout []byte, err error) {

	logger := c.logger.WithFields(map[string]interface{}{
		"env":     env,
		"command": cmd,
	})
	logger.Debug("Trying to execute with ssh")
	defer func() {
		logger.WithFields(map[string]interface{}{
			"error":  err,
			"stdout": string(stdout),
		}).Debug("Done executing command with ssh")
	}()

	var output []byte
	sess, close, err := c.open()
	defer close()
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	sess.Stdin = stdin
	sess.Stderr = &buf

	envPre := ""
	for key, value := range env {
		if key == "" || value == "" {
			if err := errors.Errorf("environment variable %s=%s is not valid", key, value); err != nil {
				return nil, err
			}
		}
		sess.Setenv(key, value)
	}
	output, err = sess.Output(envPre + cmd)
	if err != nil {
		return output, errors.Wrapf(err, "executing %s on compute %s failed with stderr %s", cmd, c.ID(), buf.String())
	}
	return output, nil
}

func (c *compute) WriteFile(path string, data io.Reader, permissions uint16) (err error) {

	logger := c.logger.WithFields(map[string]interface{}{
		"path":        path,
		"permissions": permissions,
	})
	logger.Debug("Trying to write file with ssh")
	defer func() {
		logger.WithFields(map[string]interface{}{
			"error": err,
		}).Debug("Done writing file with ssh")
	}()

	if _, err := c.Execute(nil, nil, fmt.Sprintf("sudo mkdir -p %s && sudo chown -R %s %s", filepath.Dir(path), c.remoteUser, filepath.Dir(path))); err != nil {
		return err
	}

	sess, close, err := c.open()
	defer close()
	if err != nil {
		return errors.Wrapf(err, "ssh-ing to compute %s failed", c.ID())
	}
	var stderr bytes.Buffer
	sess.Stderr = &stderr
	if logger.IsVerbose() {
		sess.Stdout = os.Stdout
	}
	sess.Stdin = data

	cmd := fmt.Sprintf("sudo sh -c 'cat > %s && chmod %d %s && chown %s %s'", path, permissions, path, c.remoteUser, path)
	if err := sess.Run(cmd); err != nil {
		return errors.Wrapf(err, "executing %s with ssh on compute %s failed with stderr %s", cmd, c.ID(), stderr.String())
	}

	return nil
}

func (c *compute) ReadFile(path string, data io.Writer) (err error) {

	logger := c.logger.WithFields(map[string]interface{}{
		"path": path,
	})
	logger.Debug("Trying to read file with ssh")
	defer func() {
		logger.WithFields(map[string]interface{}{
			"error": err,
		}).Debug("Done writing file with ssh")
	}()

	cmd := fmt.Sprintf("sudo cat %s", path)
	sess, close, err := c.open()
	defer close()
	if err != nil {
		return errors.Wrapf(err, "ssh-ing to compute %s failed", c.ID())
	}
	var stderr bytes.Buffer
	sess.Stdout = data
	sess.Stderr = &stderr

	if err := sess.Run(cmd); err != nil {
		return errors.Wrapf(err, "executing %s with ssh on compute %s failed with stderr %s", cmd, c.ID(), stderr.String())
	}
	return nil
}

func (c *compute) open() (sess *sshlib.Session, close func() error, err error) {

	c.logger.Debug("Trying to open an ssh connection")
	close = func() error { return nil }

	if c.sshCfg == nil {
		return nil, close, errors.New("no ssh key passed via infra.Compute.UseKey")
	}

	ip, err := c.compute.ExternalIP()
	if err != nil {
		return nil, close, errors.Wrap(err, "getting external IP failed")
	}

	address := fmt.Sprintf("%s:%d", *ip, 22)
	conn, err := sshlib.Dial("tcp", address, c.sshCfg)
	if err != nil {
		return nil, close, errors.Wrapf(err, "dialling tcp %s failed", address)
	}

	sess, err = conn.NewSession()
	if err != nil {
		conn.Close()
		return sess, close, err
	}
	return sess, func() error {
		err := sess.Close()
		err = conn.Close()
		return err
	}, nil
}

func (c *compute) UseKeys(sec *operator.Secrets, privateKeyPaths ...string) error {

	var signers []sshlib.Signer
	for _, privateKeyPath := range privateKeyPaths {
		privateKey, err := sec.Read(privateKeyPath)
		if err != nil {
			return err
		}

		signer, err := sshlib.ParsePrivateKey(privateKey)
		if err != nil {
			return errors.Wrap(err, "parsing private key failed")
		}
		signers = append(signers, signer)
	}

	c.sshCfg = &sshlib.ClientConfig{
		User:            c.remoteUser,
		Auth:            []sshlib.AuthMethod{sshlib.PublicKeys(signers...)},
		HostKeyCallback: sshlib.InsecureIgnoreHostKey(),
	}
	return nil
}
