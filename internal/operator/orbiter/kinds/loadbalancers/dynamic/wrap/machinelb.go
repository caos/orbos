package wrap

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
)

type cmpLB struct {
	infra.Machine
	desire func() error
}

func machine(machine infra.Machine, desire func() error) infra.Machine {
	return &cmpLB{
		Machine: machine,
		desire:  desire,
	}
}

func (c *cmpLB) Remove() error {
	err := c.Machine.Remove()
	if err != nil {
		return err
	}
	return c.desire()
}
