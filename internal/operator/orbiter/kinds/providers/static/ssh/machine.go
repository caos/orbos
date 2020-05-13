package ssh

import (
	"bytes"
	"fmt"
	"github.com/caos/orbos/internal/ssh"
	"io"
	"os"
	"path/filepath"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	sshlib "golang.org/x/crypto/ssh"
)

type ProvidedMachine interface {
	ID() string
	IP() string
	Remove() error
}

type machine struct {
	monitor    mntr.Monitor
	machine    ProvidedMachine
	remoteUser string
	sshCfg     *sshlib.ClientConfig
}

func NewMachine(monitor mntr.Monitor, comp ProvidedMachine, remoteUser string) infra.Machine {
	return &machine{
		remoteUser: remoteUser,
		monitor: monitor.WithFields(map[string]interface{}{
			"machine": comp.ID(),
		}),
		machine: comp,
	}
}

func (c *machine) ID() string {
	return c.machine.ID()
}

func (c *machine) IP() string {
	return c.machine.IP()
}

func (c *machine) Remove() error {
	return c.machine.Remove()
}

func (c *machine) Execute(env map[string]string, stdin io.Reader, cmd string) (stdout []byte, err error) {

	monitor := c.monitor.WithFields(map[string]interface{}{
		"env":     env,
		"command": cmd,
	})
	monitor.Debug("Trying to execute with ssh")
	defer func() {
		monitor.WithFields(map[string]interface{}{
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

	buf := new(bytes.Buffer)
	sess.Stdin = stdin
	sess.Stderr = buf

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
		return output, errors.Wrapf(err, "executing %s on machine %s failed with stderr %s", cmd, c.ID(), buf.String())
	}
	return output, nil
}

func (c *machine) WriteFile(path string, data io.Reader, permissions uint16) (err error) {

	monitor := c.monitor.WithFields(map[string]interface{}{
		"path":        path,
		"permissions": permissions,
	})
	monitor.Debug("Trying to write file with ssh")
	defer func() {
		monitor.WithFields(map[string]interface{}{
			"error": err,
		}).Debug("Done writing file with ssh")
	}()

	if _, err := c.Execute(nil, nil, fmt.Sprintf("sudo mkdir -p %s && sudo chown -R %s %s", filepath.Dir(path), c.remoteUser, filepath.Dir(path))); err != nil {
		return err
	}

	sess, close, err := c.open()
	defer close()
	if err != nil {
		return errors.Wrapf(err, "ssh-ing to machine %s failed", c.ID())
	}
	stderr := new(bytes.Buffer)
	sess.Stderr = stderr
	if monitor.IsVerbose() {
		sess.Stdout = os.Stdout
	}
	sess.Stdin = data

	cmd := fmt.Sprintf("sudo sh -c 'cat > %s && chmod %d %s && chown %s %s'", path, permissions, path, c.remoteUser, path)
	if err := sess.Run(cmd); err != nil {
		return errors.Wrapf(err, "executing %s with ssh on machine %s failed with stderr %s", cmd, c.ID(), stderr.String())
	}

	return nil
}

func (c *machine) ReadFile(path string, data io.Writer) (err error) {

	monitor := c.monitor.WithFields(map[string]interface{}{
		"path": path,
	})
	monitor.Debug("Trying to read file with ssh")
	defer func() {
		monitor.WithFields(map[string]interface{}{
			"error": err,
		}).Debug("Done reading file with ssh")
	}()

	cmd := fmt.Sprintf("sudo cat %s", path)
	sess, close, err := c.open()
	defer close()
	if err != nil {
		return errors.Wrapf(err, "ssh-ing to machine %s failed", c.ID())
	}
	stderr := new(bytes.Buffer)
	sess.Stdout = data
	sess.Stderr = stderr

	if err := sess.Run(cmd); err != nil {
		return errors.Wrapf(err, "executing %s with ssh on machine %s failed with stderr %s", cmd, c.ID(), stderr.String())
	}
	return nil
}

func (c *machine) open() (sess *sshlib.Session, close func() error, err error) {

	c.monitor.Debug("Trying to open an ssh connection")
	close = func() error { return nil }

	if c.sshCfg == nil {
		return nil, close, errors.New("no ssh key passed via infra.Machine.UseKey")
	}

	ip := c.machine.IP()
	address := fmt.Sprintf("%s:%d", ip, 22)

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

func (c *machine) UseKey(keys ...[]byte) error {

	publicKeys := make([]sshlib.AuthMethod, 0)
	for _, key := range keys {
		publicKey, err := ssh.PrivateKeyToPublicKey(key)
		if err != nil {
			return err
		}
		publicKeys = append(publicKeys, publicKey)
	}

	c.sshCfg = &sshlib.ClientConfig{
		User:            c.remoteUser,
		Auth:            publicKeys,
		HostKeyCallback: sshlib.InsecureIgnoreHostKey(),
	}
	return nil
}
