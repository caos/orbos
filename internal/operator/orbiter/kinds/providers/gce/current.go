package gce

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"
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

func initPools(current *Current, desired *Spec, context *context, normalized []*normalizedLoadbalancer, machines core.MachinesService) error {

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
	}
	return nil
}
