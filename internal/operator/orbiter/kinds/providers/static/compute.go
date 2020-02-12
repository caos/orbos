package static

import (
	"strings"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/static/ssh"
	"github.com/caos/orbiter/logging"
)

type machine struct {
	poolFile string
	id       *string
	ip       string
	ssh      infra.Machine
}

func newMachine(logger logging.Logger, poolFile string, remoteUser string, id *string, IP string) infra.Machine {
	cmp := &machine{poolFile: poolFile, id: id, ip: IP}
	cmp.ssh = ssh.NewMachine(logger, cmp, remoteUser)
	return cmp.ssh
}

func (c *machine) ID() string {
	return *c.id
}

func (c *machine) IP() string {
	return c.ip
}

func (c *machine) Remove() error {
	return c.ssh.WriteFile(c.poolFile, strings.NewReader(""), 600)
}
