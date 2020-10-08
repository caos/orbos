package cs

import "github.com/caos/orbos/internal/helpers"

func destroy(machinesSvc *machinesService) error {
	pools, err := machinesSvc.ListPools()
	if err != nil {
		return err
	}
	var delFuncs []func() error
	for idx := range pools {
		pool := pools[idx]
		machines, err := machinesSvc.List(pool)
		if err != nil {
			return err
		}
		for idx := range machines {
			machine := machines[idx]
			delFuncs = append(delFuncs, machine.Remove)
		}
	}
	return helpers.Fanout(delFuncs)()
}
