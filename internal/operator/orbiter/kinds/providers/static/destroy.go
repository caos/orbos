package static

import "github.com/caos/orbiter/logging"

func destroy(logger logging.Logger, desired *DesiredV0, current *Current, secrets Secrets, id string) error {
	computesSvc := NewComputesService(
		logger,
		desired,
		[]byte(secrets.BootstrapKeyPrivate.Value),
		[]byte(secrets.MaintenanceKeyPrivate.Value),
		[]byte(secrets.MaintenanceKeyPublic.Value),
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
