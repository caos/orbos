package adapter

import (
	"strings"

	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/kinds/providers/edge/ssh"
	"github.com/caos/orbiter/logging"
)

type compute struct {
	poolFile   string
	id         *string
	domainName string
	ssh        infra.Compute
}

func newCompute(logger logging.Logger, poolFile string, remoteUser string, id *string, domainName string) infra.Compute {
	cmp := &compute{poolFile: poolFile, id: id, domainName: domainName}
	cmp.ssh = ssh.NewCompute(logger, cmp, remoteUser)
	return cmp.ssh
}

func (c *compute) ID() string {
	return *c.id
}

func (c *compute) DomainName() string {
	return c.domainName
}

func (c *compute) Remove() error {
	return c.ssh.WriteFile(c.poolFile, strings.NewReader(""), 600)
}
