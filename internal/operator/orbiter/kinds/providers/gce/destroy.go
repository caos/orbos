package gce

import (
	"github.com/caos/orbiter/mntr"
	"github.com/pkg/errors"
)

func destroy(monitor mntr.Monitor, desired *Desired, current *Current, id string) error {

	return errors.New("Not yet implemented")
	/*
		machinesSvc := NewMachinesService(
			monitor,
			desired,
			[]byte(desired.Spec.Keys.BootstrapKeyPrivate.Value),
			[]byte(desired.Spec.Keys.MaintenanceKeyPrivate.Value),
			[]byte(desired.Spec.Keys.MaintenanceKeyPublic.Value),
			id,
			nil)

		pools, err := machinesSvc.ListPools()
		if err != nil {
			return err
		}

		for _, pool := range pools {
			machines, err := machinesSvc.List(pool, true)
			if err != nil {
				return err
			}
			for _, machine := range machines {
				if err := machine.Remove(); err != nil {
					return err
				}
			}
		}
		return addPools(current, desired, machinesSvc)*/
}
