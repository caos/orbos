package kubernetes

import (
	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/kubernetes/edge/k8s"
	"github.com/caos/orbiter/logging"
)

type initializedPool struct {
	infra    infra.Pool
	tier     Tier
	desired  Pool
	computes func() ([]initializedCompute, error)
}

type initializeFunc func(initializedPool, []initializedCompute) error

func (i *initializedPool) enhance(initialize initializeFunc) {
	original := i.computes
	i.computes = func() ([]initializedCompute, error) {
		computes, err := original()
		if err != nil {
			return nil, err
		}
		if err := initialize(*i, computes); err != nil {
			return nil, err
		}
		return computes, nil
	}
}

type initializedCompute struct {
	infra            infra.Compute
	tier             Tier
	currentNodeagent *common.NodeAgentCurrent
	desiredNodeagent *common.NodeAgentSpec
	markAsRunning    func()
}

func initialize(
	logger logging.Logger,
	curr *CurrentCluster,
	desired DesiredV0,
	nodeAgentsCurrent map[string]*common.NodeAgentCurrent,
	nodeAgentsDesired map[string]*common.NodeAgentSpec,
	providerPools map[string]map[string]infra.Pool) (controlplane initializedPool, workers []initializedPool, initializeCompute func(compute infra.Compute, pool initializedPool) initializedCompute, err error) {

	curr.Status = "maintaining"
	curr.Computes = make(map[string]*Compute)

	initializePool := func(infraPool infra.Pool, desired Pool, tier Tier) initializedPool {
		pool := initializedPool{
			infra:   infraPool,
			tier:    tier,
			desired: desired,
		}
		pool.computes = func() ([]initializedCompute, error) {
			infraComputes, err := infraPool.GetComputes(true)
			if err != nil {
				return nil, err
			}
			computes := make([]initializedCompute, len(infraComputes))
			for i, infraCompute := range infraComputes {
				computes[i] = initializeCompute(infraCompute, pool)
			}
			return computes, nil
		}
		return pool
	}

	initializeCompute = func(compute infra.Compute, pool initializedPool) initializedCompute {

		current := &Compute{
			Status: "maintaining",
			Metadata: ComputeMetadata{
				Tier:     pool.tier,
				Provider: pool.desired.Provider,
				Pool:     pool.desired.Pool,
			},
		}
		curr.Computes[compute.ID()] = current

		naSpec, ok := nodeAgentsDesired[compute.ID()]
		if !ok {
			naSpec = &common.NodeAgentSpec{}
			nodeAgentsDesired[compute.ID()] = naSpec
		}
		naSpec.ChangesAllowed = !pool.desired.UpdatesDisabled

		naCurr, ok := nodeAgentsCurrent[compute.ID()]
		if !ok || naCurr == nil {
			naCurr = &common.NodeAgentCurrent{}
			nodeAgentsCurrent[compute.ID()] = naCurr
		}

		if naSpec.Software == nil {
			naSpec.Software = &common.Software{}
		}

		naSpec.Software.Merge(k8s.ParseString(desired.Spec.Versions.Kubernetes).DefineSoftware())
		naSpec.Software.Merge(k8s.Current(naCurr.Software))

		return initializedCompute{
			infra: compute,
			markAsRunning: func() {
				current.Status = "running"
			},
			currentNodeagent: naCurr,
			desiredNodeagent: naSpec,
			tier:             pool.tier,
		}
	}

	for providerName, provider := range providerPools {
	pools:
		for poolName, pool := range provider {
			if desired.Spec.ControlPlane.Provider == providerName && desired.Spec.ControlPlane.Pool == poolName {
				controlplane = initializePool(pool, desired.Spec.ControlPlane, Controlplane)
				continue
			}

			for _, desiredPool := range desired.Spec.Workers {
				if providerName == desiredPool.Provider && poolName == desiredPool.Pool {
					workers = append(workers, initializePool(pool, *desiredPool, Workers))
					continue pools
				}
			}
		}
	}
	return controlplane, workers, initializeCompute, nil
}
