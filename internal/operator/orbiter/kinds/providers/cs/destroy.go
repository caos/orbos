package cs

import "github.com/caos/orbos/internal/helpers"

func destroy(machinesSvc *machinesService) error {
	pools, err := machinesSvc.ListPools()
	if err != nil {
		return err
	}
	var delFuncs []func() error
	for _, pool := range pools {
		machines, err := machinesSvc.List(pool)
		if err != nil {
			return err
		}
		for _, machine := range machines {
			delFuncs = append(delFuncs, machine.Remove)
		}
	}
	return helpers.Fanout(delFuncs)()
}
