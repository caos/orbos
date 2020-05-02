package gce

import (
	"io"
	"strings"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/static/ssh"
	"github.com/caos/orbiter/mntr"
)

var _ infra.Machine = (*machine)(nil)

type machine struct {
	active   bool
	poolFile string
	id       *string
	ip       string
	ssh      infra.Machine
}

func newMachine(monitor mntr.Monitor, poolFile string, remoteUser string, id *string, IP string) *machine {
	cmp := &machine{poolFile: poolFile, id: id, ip: IP}
	cmp.ssh = ssh.NewMachine(monitor, cmp, remoteUser)
	return cmp
}

func (c *machine) ID() string {
	return *c.id
}

func (c *machine) IP() string {
	return c.ip
}

func (c *machine) Remove() error {
	if err := c.ssh.WriteFile(c.poolFile, strings.NewReader(""), 600); err != nil {
		return err
	}
	c.active = false
	return nil
}

func (c *machine) Execute(env map[string]string, stdin io.Reader, cmd string) ([]byte, error) {
	return c.ssh.Execute(env, stdin, cmd)
}
func (c *machine) WriteFile(path string, data io.Reader, permissions uint16) error {
	return c.ssh.WriteFile(path, data, permissions)
}
func (c *machine) ReadFile(path string, data io.Writer) error {
	return c.ssh.ReadFile(path, data)
}

func (c *machine) UseKey(keys ...[]byte) error {
	return c.ssh.UseKey(keys...)
}
