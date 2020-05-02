package gce

import (
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/ssh"
	"github.com/caos/orbiter/mntr"
)

var _ infra.Machine = (*machine)(nil)

type machine struct {
	mntr.Monitor
	id     *string
	ip     string
	remove func() error
	*ssh.Machine
}

func newMachine(monitor mntr.Monitor, id *string, IP string, remove func() error) *machine {
	return &machine{
		Monitor: monitor,
		id:      id,
		ip:      IP,
		remove:  remove,
		Machine: ssh.NewMachine(monitor, IP, "orbiter"),
	}
}

func (c *machine) ID() string {
	return *c.id
}

func (c *machine) IP() string {
	return c.ip
}

func (c *machine) Remove() error {
	return c.remove()
}
