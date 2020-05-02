package gce

import (
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/ssh"
	"github.com/caos/orbiter/mntr"
)

var _ infra.Machine = (*machine)(nil)

type machine struct {
	mntr.Monitor
	id     string
	ip     string
	pool   string
	remove func() error
	*ssh.Machine
}

func newMachine(monitor mntr.Monitor, id, ip, pool string, remove func() error) *machine {
	return &machine{
		Monitor: monitor,
		id:      id,
		ip:      ip,
		pool:    pool,
		remove:  remove,
		Machine: ssh.NewMachine(monitor, ip, "orbiter"),
	}
}

func (c *machine) ID() string {
	return c.id
}

func (c *machine) IP() string {
	return c.ip
}

func (c *machine) Remove() error {
	return c.remove()
}
