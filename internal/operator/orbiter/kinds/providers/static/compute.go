package static

import (
	"strings"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/static/ssh"
	"github.com/caos/orbiter/logging"
)

type compute struct {
	poolFile string
	id       *string
	ip       string
	ssh      infra.Compute
}

func newCompute(logger logging.Logger, poolFile string, remoteUser string, id *string, IP string) infra.Compute {
	cmp := &compute{poolFile: poolFile, id: id, ip: IP}
	cmp.ssh = ssh.NewCompute(logger, cmp, remoteUser)
	return cmp.ssh
}

func (c *compute) ID() string {
	return *c.id
}

func (c *compute) IP() string {
	return c.ip
}

func (c *compute) Remove() error {
	return c.ssh.WriteFile(c.poolFile, strings.NewReader(""), 600)
}
