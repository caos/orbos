package ssh

import (
	"bytes"
	"fmt"
	"io"
	"path/filepath"

	"github.com/caos/orbos/internal/ssh"

	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"

	sshlib "golang.org/x/crypto/ssh"
)

type Machine struct {
	monitor    mntr.Monitor
	remoteUser string
	ip         string
	sshCfg     *sshlib.ClientConfig
}

func NewMachine(monitor mntr.Monitor, remoteUser, ip string) *Machine {
	return &Machine{
		remoteUser: remoteUser,
		monitor: monitor.WithFields(map[string]interface{}{
			"host": ip,
			"user": remoteUser,
		}),
		ip: ip,
	}
}

func (c *Machine) Execute(env map[string]string, stdin io.Reader, cmd string) (stdout []byte, err error) {

	monitor := c.monitor.WithFields(map[string]interface{}{
		"env":     env,
		"command": cmd,
	})
	defer func() {
		if err != nil {
			err = fmt.Errorf("executing %s failed: %w", cmd, err)
		} else {
			monitor.WithField("stdout", string(stdout)).Debug("Done executing command with ssh")
		}
	}()

	monitor.Debug("Trying to execute with ssh")

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
			return nil, errors.Errorf("environment variable %s=%s is not valid", key, value)
		}
		sess.Setenv(key, value)
	}
	output, err = sess.Output(envPre + cmd)
	if err != nil {
		return output, fmt.Errorf("stderr: %s", buf.String())
	}
	return output, nil
}

func (c *Machine) WriteFile(path string, data io.Reader, permissions uint16) (err error) {

	monitor := c.monitor.WithFields(map[string]interface{}{
		"path":        path,
		"permissions": permissions,
	})
	defer func() {
		if err != nil {
			err = fmt.Errorf("writing file %s failed: %w", path, err)
		} else {
			monitor.Debug("Done writing file with ssh")
		}
	}()

	monitor.Debug("Trying to write file with ssh")

	if _, err := c.Execute(nil, nil, fmt.Sprintf("sudo mkdir -p %s && sudo chown -R %s %s", filepath.Dir(path), c.remoteUser, filepath.Dir(path))); err != nil {
		return err
	}

	_, err = c.Execute(nil, data, fmt.Sprintf("sudo sh -c 'cat > %s && chmod %d %s && chown %s %s'", path, permissions, path, c.remoteUser, path))
	return err
}

func (c *Machine) ReadFile(path string, data io.Writer) (err error) {

	monitor := c.monitor.WithFields(map[string]interface{}{
		"path": path,
	})
	defer func() {
		if err != nil {
			err = fmt.Errorf("reading file %s failed: %w", path, err)
		} else {
			monitor.Debug("Done reading file with ssh")
		}
	}()

	monitor.Debug("Trying to read file with ssh")

	cmd := fmt.Sprintf("sudo cat %s", path)
	sess, close, err := c.open()
	defer close()
	if err != nil {
		return err
	}
	stderr := new(bytes.Buffer)
	sess.Stdout = data
	sess.Stderr = stderr

	if err := sess.Run(cmd); err != nil {
		return fmt.Errorf("executing %s failed with stderr %s: %w", cmd, stderr.String(), err)
	}
	return nil
}

func (c *Machine) open() (sess *sshlib.Session, close func() error, err error) {

	c.monitor.Debug("Trying to open an ssh connection")
	close = func() error { return nil }

	if c.sshCfg == nil {
		return nil, close, errors.New("no ssh key passed via infra.Machine.UseKey")
	}

	address := fmt.Sprintf("%s:%d", c.ip, 22)
	conn, err := sshlib.Dial("tcp", address, c.sshCfg)
	if err != nil {
		return nil, close, errors.Wrapf(err, "dialling tcp %s with user %s failed", address, c.remoteUser)
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

func (c *Machine) UseKey(keys ...[]byte) error {

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
