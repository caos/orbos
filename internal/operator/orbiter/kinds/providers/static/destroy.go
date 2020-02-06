package static

import "github.com/caos/orbiter/logging"

func destroy(logger logging.Logger, desired *DesiredV0, current *Current, id string) error {
	computesSvc := NewComputesService(
		logger,
		desired,
		[]byte(desired.Spec.Keys.BootstrapKeyPrivate.Value),
		[]byte(desired.Spec.Keys.MaintenanceKeyPrivate.Value),
		[]byte(desired.Spec.Keys.MaintenanceKeyPublic.Value),
		id,
		nil)

	pools, err := computesSvc.ListPools()
	if err != nil {
		return err
	}

	for _, pool := range pools {
		computes, err := computesSvc.List(pool, true)
		if err != nil {
			return err
		}
		for _, compute := range computes {
			if err := compute.Remove(); err != nil {
				return err
			}
		}
	}
	return addPools(current, desired, computesSvc)
}
