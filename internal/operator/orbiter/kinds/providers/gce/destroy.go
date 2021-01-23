package gce

import "github.com/caos/orbos/internal/helpers"

func destroy(svc *machinesService) error {
	return helpers.Fanout([]func() error{
		func() error {
			destroyLB, err := queryLB(svc.context, nil)
			if err != nil {
				return err
			}
			return destroyLB()
		},
		func() error {
			pools, err := svc.ListPools()
			if err != nil {
				return err
			}
			var delFuncs []func() error
			for _, pool := range pools {
				machines, err := svc.List(pool)
				if err != nil {
					return err
				}
				for _, machine := range machines {
					delFuncs = append(delFuncs, machine.Remove)
				}
			}
			if err := helpers.Fanout(delFuncs)(); err != nil {
				return err
			}
			_, deleteFirewalls, err := queryFirewall(svc.context, nil)
			if err != nil {
				return err
			}
			return destroyNetwork(svc.context, deleteFirewalls)
		},
	})()
}
