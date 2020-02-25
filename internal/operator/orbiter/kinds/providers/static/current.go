package static

import (
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/core"
)

func addPools(current *Current, desired *DesiredV0, machinesSvc core.MachinesService) error {
	current.Current.pools = make(map[string]infra.Pool)
	for pool := range desired.Spec.Pools {
		current.Current.pools[pool] = core.NewPool(pool, nil, machinesSvc)
	}

	unconfiguredPools, err := machinesSvc.ListPools()
	if err != nil {
		return nil
	}
	for _, unconfiguredPool := range unconfiguredPools {
		if _, ok := current.Current.pools[unconfiguredPool]; !ok {
			current.Current.pools[unconfiguredPool] = core.NewPool(unconfiguredPool, nil, machinesSvc)
		}
	}
	return nil
}
