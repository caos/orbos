package gce

import "github.com/caos/orbos/internal/helpers"

func destroy(context *context) error {
	if err := helpers.Fanout([]func() error{
		func() error {
			destroyLB, err := queryLB(context, nil)
			if err != nil {
				return err
			}
			return destroyLB()
		},
		func() error {
			pools, err := context.machinesService.ListPools()
			if err != nil {
				return err
			}
			var delFuncs []func() error
			for _, pool := range pools {
				machines, err := context.machinesService.List(pool)
				if err != nil {
					return err
				}
				for _, machine := range machines {
					delFuncs = append(delFuncs, machine.Remove)
				}
			}
			return helpers.Fanout(delFuncs)()
		},
		func() error {
			return destroyNetwork(context)
		},
	})(); err != nil {
		return err
	}
	return destroyNetwork(context)
}
