package kubernetes

import (
	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/logging"
)

type initializedPool struct {
	infra    infra.Pool
	tier     Tier
	desired  Pool
	machines func() ([]initializedMachine, error)
}

type initializeFunc func(initializedPool, []initializedMachine) error

func (i *initializedPool) enhance(initialize initializeFunc) {
	original := i.machines
	i.machines = func() ([]initializedMachine, error) {
		machines, err := original()
		if err != nil {
			return nil, err
		}
		if err := initialize(*i, machines); err != nil {
			return nil, err
		}
		return machines, nil
	}
}

type initializedMachine struct {
	infra            infra.Machine
	tier             Tier
	currentNodeagent *common.NodeAgentCurrent
	desiredNodeagent *common.NodeAgentSpec
	currentMachine   *Machine
}

func initialize(
	logger logging.Logger,
	curr *CurrentCluster,
	desired DesiredV0,
	nodeAgentsCurrent map[string]*common.NodeAgentCurrent,
	nodeAgentsDesired map[string]*common.NodeAgentSpec,
	providerPools map[string]map[string]infra.Pool) (controlplane initializedPool, workers []initializedPool, initializeMachine func(machine infra.Machine, pool initializedPool) initializedMachine, err error) {

	curr.Status = "maintaining"
	curr.Machines = make(map[string]*Machine)

	initializePool := func(infraPool infra.Pool, desired Pool, tier Tier) initializedPool {
		pool := initializedPool{
			infra:   infraPool,
			tier:    tier,
			desired: desired,
		}
		pool.machines = func() ([]initializedMachine, error) {
			infraMachines, err := infraPool.GetMachines(true)
			if err != nil {
				return nil, err
			}
			machines := make([]initializedMachine, len(infraMachines))
			for i, infraMachine := range infraMachines {
				machines[i] = initializeMachine(infraMachine, pool)
			}
			return machines, nil
		}
		return pool
	}

	initializeMachine = func(machine infra.Machine, pool initializedPool) initializedMachine {

		current := &Machine{
			Node: NodeStatus{
				Joined:      false,
				Online:      false,
				Maintaining: true,
			},
			Metadata: MachineMetadata{
				Tier:     pool.tier,
				Provider: pool.desired.Provider,
				Pool:     pool.desired.Pool,
			},
		}
		curr.Machines[machine.ID()] = current

		naSpec, ok := nodeAgentsDesired[machine.ID()]
		if !ok {
			naSpec = &common.NodeAgentSpec{}
			nodeAgentsDesired[machine.ID()] = naSpec
		}
		naSpec.ChangesAllowed = !pool.desired.UpdatesDisabled

		naCurr, ok := nodeAgentsCurrent[machine.ID()]
		if !ok || naCurr == nil {
			naCurr = &common.NodeAgentCurrent{}
			nodeAgentsCurrent[machine.ID()] = naCurr
		}

		if naSpec.Software == nil {
			naSpec.Software = &common.Software{}
		}

		naSpec.Software.Merge(ParseString(desired.Spec.Versions.Kubernetes).DefineSoftware())
		naSpec.Software.Merge(KubernetesSoftware(naCurr.Software))

		return initializedMachine{
			infra:            machine,
			currentNodeagent: naCurr,
			desiredNodeagent: naSpec,
			tier:             pool.tier,
			currentMachine:   current,
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
	return controlplane, workers, initializeMachine, nil
}
