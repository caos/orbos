package static

import (
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/core"
)

func addPools(current *Current, desired *DesiredV0, computesSvc core.ComputesService) error {
	current.Current.Pools = make(map[string]infra.Pool)
	for pool := range desired.Spec.Pools {
		current.Current.Pools[pool] = core.NewPool(pool, nil, computesSvc)
	}

	unconfiguredPools, err := computesSvc.ListPools()
	if err != nil {
		return nil
	}
	for _, unconfiguredPool := range unconfiguredPools {
		if _, ok := current.Current.Pools[unconfiguredPool]; !ok {
			current.Current.Pools[unconfiguredPool] = core.NewPool(unconfiguredPool, nil, computesSvc)
		}
	}
	return nil
}
