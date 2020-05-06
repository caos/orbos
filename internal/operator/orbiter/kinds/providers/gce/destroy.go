package gce

import (
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/core"
)

func destroy(machinesService core.MachinesService) error {
	pools, err := machinesService.ListPools()
	if err != nil {
		return err
	}
	for _, pool := range pools {
		machines, err := machinesService.List(pool)
		if err != nil {
			return err
		}
		for _, machine := range machines {
			if err := machine.Remove(); err != nil {
				return err
			}
		}
	}
	return nil
}
