package static

import "github.com/caos/orbos/mntr"

func destroy(monitor mntr.Monitor, desired *DesiredV0, current *Current, id string) error {

	machinesSvc := NewMachinesService(
		monitor,
		desired,
		id,
		nil)

	if err := machinesSvc.updateKeys(); err != nil {
		return err
	}

	pools, err := machinesSvc.ListPools()
	if err != nil {
		return err
	}

	for _, pool := range pools {
		machines, err := machinesSvc.List(pool)
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
