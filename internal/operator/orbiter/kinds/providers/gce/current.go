package gce

import (
	"fmt"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic/wrap"
	"github.com/caos/orbos/internal/tree"
)

type Current struct {
	Common  *tree.Common `yaml:",inline"`
	Current struct {
		pools      map[string]infra.Pool `yaml:"-"`
		Ingresses  map[string]*infra.Address
		cleanupped <-chan error `yaml:"-"`
	}
}

func (c *Current) Pools() map[string]infra.Pool {
	return c.Current.pools
}
func (c *Current) Ingresses() map[string]*infra.Address {
	return c.Current.Ingresses
}
func (c *Current) Cleanupped() <-chan error {
	return c.Current.cleanupped
}

func initPools(current *Current, desired *Spec, context *context, normalized []*normalizedLoadbalancer, lbCurrent *dynamic.Current, nodeAgentsDesired map[string]*common.NodeAgentSpec) error {

	mapVipFunc := func(vip *dynamic.VIP) string {
		for _, transport := range vip.Transport {
			address, ok := current.Current.Ingresses[transport.Name]
			if ok {
				return address.Location
			}
		}
		panic(fmt.Errorf("external address for %v is not ensured", vip))
	}

	machines := wrap.MachinesService(context.machinesService, *lbCurrent, nodeAgentsDesired, false, nil, mapVipFunc)

	current.Current.pools = make(map[string]infra.Pool)
	for pool := range desired.Pools {
		current.Current.pools[pool] = newInfraPool(pool, context, normalized, machines)
	}

	pools, err := machines.ListPools()
	if err != nil {
		return nil
	}
	for _, pool := range pools {
		// Also return pools that are not configured
		if _, ok := current.Current.pools[pool]; !ok {
			current.Current.pools[pool] = newInfraPool(pool, context, normalized, machines)
		}
		// initialize existing machines
		lbCurrent.Current.Desire(pool, context.machinesService, nodeAgentsDesired, false, nil, mapVipFunc)
		machines, err := machines.List(pool)
		if err != nil {
			return err
		}
		for _, machine := range machines {
			context.machinesService.onCreate(pool, machine)
		}
	}
	return nil
}
