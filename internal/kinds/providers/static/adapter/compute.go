package adapter

import (
	"strings"

	"github.com/caos/orbiter/logging"
	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/kinds/providers/edge/ssh"
)

type compute struct {
	poolFile string
	id       *string
	intIP    *string
	extIP    *string
	ssh      infra.Compute
}

func newCompute(logger logging.Logger, poolFile string, remoteUser string, id *string, intIP *string, extIP *string) infra.Compute {
	cmp := &compute{poolFile: poolFile, id: id, intIP: intIP, extIP: extIP}
	cmp.ssh = ssh.NewCompute(logger, cmp, remoteUser)
	return cmp.ssh
}

func (c *compute) ID() string {
	return *c.id
}

func (c *compute) InternalIP() (*string, error) {
	return c.intIP, nil

}

func (c *compute) ExternalIP() (*string, error) {
	return c.extIP, nil
}

func (c *compute) Remove() error {
	return c.ssh.WriteFile(c.poolFile, strings.NewReader(""), 600)
}
