package static

import (
	"strings"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/ssh"
	"github.com/caos/orbos/mntr"
)

var _ infra.Machine = (*machine)(nil)

type machine struct {
	active   bool
	poolFile string
	id       *string
	ip       string
	*ssh.Machine
}

func newMachine(monitor mntr.Monitor, poolFile string, remoteUser string, id *string, ip string) *machine {
	return &machine{
		active:   false,
		poolFile: poolFile,
		id:       id,
		ip:       ip,
		Machine:  ssh.NewMachine(monitor, remoteUser, ip),
	}
}

func (c *machine) ID() string {
	return *c.id
}

func (c *machine) IP() string {
	return c.ip
}

func (c *machine) Remove() error {
	if err := c.Machine.WriteFile(c.poolFile, strings.NewReader(""), 600); err != nil {
		return err
	}
	c.active = false
	c.Execute(nil, nil, "sudo systemctl stop node-agentd")
	c.Execute(nil, nil, "sudo systemctl disable node-agentd")
	c.Execute(nil, nil, "sudo kubeadm reset -f")
	c.Execute(nil, nil, "sudo rm -rf /var/lib/etcd")
	return nil
}
