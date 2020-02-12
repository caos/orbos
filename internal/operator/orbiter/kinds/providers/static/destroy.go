package static

import "github.com/caos/orbiter/logging"

func destroy(logger logging.Logger, desired *DesiredV0, current *Current, id string) error {
	machinesSvc := NewMachinesService(
		logger,
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
	return addPools(current, desired, machinesSvc)
}
