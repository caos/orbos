package kubernetes

import (
	"fmt"
	"strings"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/mntr"
	core "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	macherrs "k8s.io/apimachinery/pkg/api/errors"
)

type initializedPool struct {
	infra    infra.Pool
	tier     Tier
	desired  Pool
	machines func() ([]*initializedMachine, error)
}

type initializeFunc func(initializedPool, []*initializedMachine) error
type uninitializeMachineFunc func(id string)
type initializeMachineFunc func(machine infra.Machine, pool initializedPool) *initializedMachine

func (i *initializedPool) enhance(initialize initializeFunc) {
	original := i.machines
	i.machines = func() ([]*initializedMachine, error) {
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
	reconcile        func() error
	currentNodeagent *common.NodeAgentCurrent
	desiredNodeagent *common.NodeAgentSpec
	currentMachine   *Machine
	pool             *initializedPool
}

func initialize(
	monitor mntr.Monitor,
	curr *CurrentCluster,
	desired DesiredV0,
	nodeAgentsCurrent *common.CurrentNodeAgents,
	nodeAgentsDesired *common.DesiredNodeAgents,
	providerPools map[string]map[string]infra.Pool,
	k8s *Client,
	postInit func(machine *initializedMachine)) (
	controlplane initializedPool,
	controlplaneMachines []*initializedMachine,
	workers []initializedPool,
	workerMachines []*initializedMachine,
	initializeMachine initializeMachineFunc,
	uninitializeMachine uninitializeMachineFunc,
	err error) {

	curr.Status = "running"

	initializePool := func(infraPool infra.Pool, desired Pool, tier Tier) initializedPool {
		pool := initializedPool{
			infra:   infraPool,
			tier:    tier,
			desired: desired,
		}
		pool.machines = func() ([]*initializedMachine, error) {
			infraMachines, err := infraPool.GetMachines()
			if err != nil {
				return nil, err
			}
			machines := make([]*initializedMachine, len(infraMachines))
			for i, infraMachine := range infraMachines {
				machines[i] = initializeMachine(infraMachine, pool)
				if !machines[i].currentMachine.Online {
					curr.Status = "maintaining"
				}
			}
			return machines, nil
		}
		return pool
	}

	initializeMachine = func(machine infra.Machine, pool initializedPool) *initializedMachine {

		current := &Machine{
			Metadata: MachineMetadata{
				Tier:     pool.tier,
				Provider: pool.desired.Provider,
				Pool:     pool.desired.Pool,
			},
		}

		var node *v1.Node
		if k8s.Available() {
			var k8sNodeErr error
			node, k8sNodeErr = k8s.GetNode(machine.ID())
			if k8sNodeErr != nil && !macherrs.IsNotFound(k8sNodeErr) {
				current.Unknown = true
			}
		}

		// Retry if kubeapi returns other error than "NotFound"

		reconcile := func() error { return nil }
		if node != nil && !current.Unknown {
			reconcile = reconcileNodeFunc(*node, monitor, pool.desired, k8s, pool.tier)
			current.Joined = true
			for _, cond := range node.Status.Conditions {
				if cond.Type == v1.NodeReady {
					current.Ready = true
					if !node.Spec.Unschedulable {
						current.Online = true
					}
				}
			}
		}

		curr.Machines.Set(machine.ID(), current)

		machineMonitor := monitor.WithField("machine", machine.ID())

		naSpec, _ := nodeAgentsDesired.Get(machine.ID())
		naSpec.ChangesAllowed = !pool.desired.UpdatesDisabled
		naCurr, _ := nodeAgentsCurrent.Get(machine.ID())
		k8sSoftware := ParseString(desired.Spec.Versions.Kubernetes).DefineSoftware()
		if !softwareDefines(*naSpec.Software, k8sSoftware) {
			k8sSoftware.Merge(KubernetesSoftware(naCurr.Software))
			if !softwareContains(*naSpec.Software, k8sSoftware) {
				naSpec.Software.Merge(k8sSoftware)
				machineMonitor.Changed("Kubernetes software desired")
			}
		}

		initMachine := &initializedMachine{
			infra:            machine,
			currentNodeagent: naCurr,
			desiredNodeagent: naSpec,
			reconcile:        reconcile,
			currentMachine:   current,
			pool:             &pool,
		}

		postInit(initMachine)

		return initMachine
	}

	for providerName, provider := range providerPools {
	pools:
		for poolName, pool := range provider {
			if desired.Spec.ControlPlane.Provider == providerName && desired.Spec.ControlPlane.Pool == poolName {
				controlplane = initializePool(pool, desired.Spec.ControlPlane, Controlplane)
				controlplaneMachines, err = controlplane.machines()
				if err != nil {
					return controlplane,
						controlplaneMachines,
						workers,
						workerMachines,
						initializeMachine,
						uninitializeMachine,
						err
				}
				continue
			}

			for _, desiredPool := range desired.Spec.Workers {
				if providerName == desiredPool.Provider && poolName == desiredPool.Pool {
					workerPool := initializePool(pool, *desiredPool, Workers)
					workers = append(workers, workerPool)
					initializedWorkerMachines, err := workerPool.machines()
					if err != nil {
						return controlplane,
							controlplaneMachines,
							workers,
							workerMachines,
							initializeMachine,
							uninitializeMachine,
							err
					}
					workerMachines = append(workerMachines, initializedWorkerMachines...)
					continue pools
				}
			}
		}
	}

	for _, machine := range append(controlplaneMachines, workerMachines...) {
		if !machine.currentMachine.Ready {
			curr.Status = "degraded"
		}
		if !machine.currentMachine.Online || !machine.currentMachine.Joined || !machine.currentMachine.FirewallIsReady {
			curr.Status = "maintaining"
			break
		}
	}

	return controlplane,
		controlplaneMachines,
		workers,
		workerMachines,
		initializeMachine, func(id string) {
			nodeAgentsDesired.Delete(id)
			curr.Machines.Delete(id)
		}, nil
}

func reconcileNodeFunc(node v1.Node, monitor mntr.Monitor, pool Pool, k8s *Client, tier Tier) func() error {
	reconcileNode := false
	reconcileMonitor := monitor.WithField("node", node.Name)
	handleMaybe := func(maybeMonitorFields map[string]interface{}) {
		if maybeMonitorFields != nil {
			reconcileNode = true
			monitor = monitor.WithFields(maybeMonitorFields)
		}
	}

	handleMaybe(reconcileLabels(&node, "orbos.ch/pool", pool.Pool))
	handleMaybe(reconcileLabels(&node, "orbos.ch/tier", string(tier)))
	handleMaybe(reconcileTaints(&node, pool))

	if !reconcileNode {
		return func() error { return nil }
	}
	return func() error {
		reconcileMonitor.Info("Reconciling node")
		return k8s.updateNode(&node)
	}
}

func reconcileTaints(node *v1.Node, pool Pool) map[string]interface{} {
	desiredTaints := pool.Taints.ToK8sTaints()
	newTaints := append([]core.Taint{}, desiredTaints...)
	updateTaints := false
outer:
	for _, existing := range node.Spec.Taints {
		if strings.HasPrefix(existing.Key, "node.kubernetes.io/") {
			newTaints = append(newTaints, existing)
			continue
		}
		for _, des := range desiredTaints {
			if existing.Key == des.Key &&
				existing.Effect == des.Effect &&
				existing.Value == des.Value {
				continue outer
			}
		}
		updateTaints = true
		break
	}
	if !updateTaints && len(node.Spec.Taints) == len(newTaints) || pool.Taints == nil {
		return nil
	}
	node.Spec.Taints = newTaints
	return map[string]interface{}{"taints": desiredTaints}
}

func reconcileLabels(node *v1.Node, key, value string) map[string]interface{} {
	if node.Labels[key] == value {
		return nil
	}
	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}
	node.Labels[key] = value
	return map[string]interface{}{
		fmt.Sprintf("label.%s", key): value,
	}
}
