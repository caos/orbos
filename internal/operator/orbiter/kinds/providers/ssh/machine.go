package ssh

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh/knownhosts"

	sshlib "golang.org/x/crypto/ssh"

	"github.com/caos/orbos/internal/ssh"
	"github.com/caos/orbos/mntr"
)

type Machine struct {
	monitor    mntr.Monitor
	remoteUser string
	ip         string
	zone       string
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

func (c *Machine) Zone() string {
	return c.zone
}

func (c *Machine) Execute(stdin io.Reader, cmd string) (stdout []byte, err error) {

	monitor := c.monitor.WithFields(map[string]interface{}{
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
	sess, closeSession, err := c.open(c.sshCfg)
	defer mustClose(closeSession)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	defer buf.Reset()
	sess.Stdin = stdin
	sess.Stderr = buf

	output, err = sess.Output(cmd)
	if err != nil {
		return output, fmt.Errorf("stderr: %s", buf.String())
	}
	return output, nil
}

func (c *Machine) Shell() (err error) {
	defer func() {
		if err != nil {
			err = fmt.Errorf("executing shell failed: %w", err)
		} else {
			c.monitor.Debug("Done executing shell with ssh")
		}
	}()

	sess, closeSession, err := c.open(c.sshCfg)
	defer mustClose(closeSession)

	if err != nil {
		return err
	}
	sess.Stdin = os.Stdin
	sess.Stderr = os.Stderr
	sess.Stdout = os.Stdout
	modes := sshlib.TerminalModes{
		sshlib.ECHO:          0,     // disable echoing
		sshlib.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		sshlib.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := sess.RequestPty("xterm", 40, 80, modes); err != nil {
		return fmt.Errorf("request for pseudo terminal failed: %w", err)
	}

	if err := sess.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}
	return sess.Wait()
}

func WriteFileCommands(user, path string, permissions uint16) (string, string) {
	return fmt.Sprintf("sudo mkdir -p %s && sudo chown -R %s %s", filepath.Dir(path), user, filepath.Dir(path)),
		fmt.Sprintf("sudo sh -c 'cat > %s && chmod %d %s && chown %s %s'", path, permissions, path, user, path)
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

	ensurePath, writeFile := WriteFileCommands(c.remoteUser, path, permissions)

	if _, err := c.Execute(nil, ensurePath); err != nil {
		return err
	}

	_, err = c.Execute(data, writeFile)
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
	sess, closeSession, err := c.open(c.sshCfg)
	defer mustClose(closeSession)

	if err != nil {
		return err
	}
	stderr := new(bytes.Buffer)
	defer stderr.Reset()
	sess.Stdout = data
	sess.Stderr = stderr

	if err := sess.Run(cmd); err != nil {
		return fmt.Errorf("executing %s failed with stderr %s: %w", cmd, stderr.String(), err)
	}
	return nil
}

func (c *Machine) open(sshCfg *sshlib.ClientConfig) (sess *sshlib.Session, close func() error, err error) {

	c.monitor.Debug("Trying to open an ssh connection")
	close = func() error { return nil }

	if sshCfg == nil {
		return nil, close, errors.New("no ssh key passed via infra.Machine.UseKey")
	}

	address := fmt.Sprintf("%s:%d", c.ip, 22)
	conn, err := sshlib.Dial("tcp", address, c.sshCfg)
	if err != nil {
		return nil, close, fmt.Errorf("dialling tcp %s with user %s failed: %w", address, c.remoteUser, err)
	}

	sess, err = conn.NewSession()
	if err != nil {
		if err := conn.Close(); err != nil {
			panic(err)
		}
		return sess, close, err
	}
	return sess, func() error {
		err := sess.Close()
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		err = conn.Close()
		if err != nil && !errors.Is(err, io.EOF) {
			return err
		}
		return nil
	}, nil
}

func (c *Machine) UseKey(keys ...[]byte) error {

	var typedCheckErr *knownhosts.KeyError

	khPath, err := ensureKnownHostsPath()
	if err != nil {
		return err
	}

	checkHost, err := knownhosts.New(khPath)
	if err != nil {
		return err
	}

	publicKeys, err := ssh.AuthMethodFromKeys(keys...)
	if err != nil {
		return err
	}

	var startedHostKeyInitialization bool
	hostKeyInitialized := make(chan struct{})

	var calls int

	c.sshCfg = &sshlib.ClientConfig{
		User: c.remoteUser,
		Auth: []sshlib.AuthMethod{publicKeys},
		HostKeyCallback: func(hostname string, remote net.Addr, key sshlib.PublicKey) (err error) {

			calls++
			fmt.Println("calls", calls)

			awaitRemoteSSHDaemon := func() error {
				sess, closeSession, err := c.open(&sshlib.ClientConfig{
					User:            c.remoteUser,
					Auth:            []sshlib.AuthMethod{publicKeys},
					HostKeyCallback: sshlib.FixedHostKey(key),
				})
				defer mustClose(closeSession)

				if err != nil {
					return err
				}
				fmt.Println("trying whoami")
				out, err := sess.Output("whoami")
				if err != nil {
					return fmt.Errorf("first time connection to the machine failed with output: %s: %w", string(out), err)
				}
				fmt.Println("executed whoami")
				if string(out) != c.remoteUser {
					return fmt.Errorf("expected the remote command whoami to output %s, but got %s", c.remoteUser, string(out))
				}
				fmt.Println("checked whoami")
				return nil
			}

			doInit := !startedHostKeyInitialization // should be an independent copy of startedHostKeyInitialization
			if doInit {
				defer func() {
					if err != nil {
						// try again
						fmt.Println("trying again")
						startedHostKeyInitialization = false
						return
					}
					err = awaitRemoteSSHDaemon()
					if err == nil {
						fmt.Println("closing hostKeyInitialized")
						close(hostKeyInitialized)
					}
				}()
			} else {
				fmt.Println("<-hostKeyInitialized waiting")
				<-hostKeyInitialized
				fmt.Println("<-hostKeyInitialized received")
			}

			startedHostKeyInitialization = true

			checkErr := checkHost(hostname, remote, key)
			if checkErr == nil {
				return nil
			}
			if !errors.As(checkErr, &typedCheckErr) {
				return checkErr
			}
			// Reference: https://www.godoc.org/golang.org/x/crypto/ssh/knownhosts#KeyError
			// if keyErr.Want slice is empty then host is unknown, if keyErr.Want is not empty
			// and if host is known then there is key mismatch the connection is then rejected.
			if len(typedCheckErr.Want) > 0 {
				return fmt.Errorf("%v is not a key of %s, either you are a victim of a MiTM attack or %s has reconfigured the host pub key: %w", string(key.Marshal()), hostname, hostname, checkErr)
			}
			c.monitor.Info("Adding missing host key to known_hosts file")
			return addHostKey(khPath, remote, key)
		},
	}
	return nil
}

func addHostKey(knownHostsPath string, remote net.Addr, pubKey sshlib.PublicKey) error {

	f, fErr := os.OpenFile(knownHostsPath, os.O_APPEND|os.O_WRONLY, 0600)
	if fErr != nil {
		return fErr
	}
	defer f.Close()

	knownHosts := knownhosts.Normalize(remote.String())
	_, fileErr := f.WriteString(knownhosts.Line([]string{knownHosts}, pubKey) + "\n")
	return fileErr
}

func ensureKnownHostsPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	sshPath := filepath.Join(home, ".ssh")
	khPath := filepath.Join(sshPath, "known_hosts")

	if err := os.MkdirAll(sshPath, 0600); err != nil {
		return "", err
	}

	f, err := os.OpenFile(khPath, os.O_CREATE, 0600)
	if err != nil {
		return "", err
	}
	return khPath, f.Close()
}

func mustClose(closeFn func() error) {
	if err := closeFn(); err != nil {
		panic(fmt.Sprintf("closing failed: %w", err))
	}
}
