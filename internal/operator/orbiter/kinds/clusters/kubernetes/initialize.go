package kubernetes

import (
	"fmt"
	"strings"
	"time"

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
	tier             Tier
	reconcile        func() error
	currentNodeagent *common.NodeAgentCurrent
	desiredNodeagent *common.NodeAgentSpec
	currentMachine   *Machine
}

func initialize(
	monitor mntr.Monitor,
	curr *CurrentCluster,
	desired DesiredV0,
	nodeAgentsCurrent map[string]*common.NodeAgentCurrent,
	nodeAgentsDesired map[string]*common.NodeAgentSpec,
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

	if curr.Machines == nil {
		curr.Machines = make(map[string]*Machine)
	}

	curr.Status = "running"

	initializePool := func(infraPool infra.Pool, desired Pool, tier Tier) initializedPool {
		pool := initializedPool{
			infra:   infraPool,
			tier:    tier,
			desired: desired,
		}
		pool.machines = func() ([]*initializedMachine, error) {
			infraMachines, err := infraPool.GetMachines(true)
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

		node, imErr := k8s.GetNode(machine.ID())

		// Retry if kubeapi returns other error than "NotFound"
		for k8s.Available() && imErr != nil && !macherrs.IsNotFound(imErr) {
			monitor.WithFields(map[string]interface{}{
				"node":  machine.ID(),
				"error": imErr.Error(),
			}).Info("Could not determine node state")
			time.Sleep(5 * time.Second)
			node, imErr = k8s.GetNode(machine.ID())
		}

		current := &Machine{
			Metadata: MachineMetadata{
				Tier:     pool.tier,
				Provider: pool.desired.Provider,
				Pool:     pool.desired.Pool,
			},
		}

		reconcile := func() error { return nil }
		if imErr == nil {
			reconcile = reconcileNodeFunc(*node, monitor, pool.desired, k8s)
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

		curr.Machines[machine.ID()] = current

		machineMonitor := monitor.WithField("machine", machine.ID())

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

		k8sSoftware := ParseString(desired.Spec.Versions.Kubernetes).DefineSoftware()
		if !naSpec.Software.Defines(k8sSoftware) {
			k8sSoftware.Merge(KubernetesSoftware(naCurr.Software))
			if !naSpec.Software.Contains(k8sSoftware) {
				naSpec.Software.Merge(k8sSoftware)
				machineMonitor.Changed("Kubernetes software desired")
			}
		}

		initMachine := &initializedMachine{
			infra:            machine,
			currentNodeagent: naCurr,
			desiredNodeagent: naSpec,
			tier:             pool.tier,
			reconcile:        reconcile,
			currentMachine:   current,
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
		if !machine.currentMachine.Online || !machine.currentMachine.Joined || !machine.currentMachine.NodeAgentIsRunning || !machine.currentMachine.FirewallIsReady {
			curr.Status = "maintaining"
			break
		}
	}

	return controlplane,
		controlplaneMachines,
		workers,
		workerMachines,
		initializeMachine, func(id string) {
			delete(nodeAgentsDesired, id)
			delete(curr.Machines, id)
		}, nil
}

func reconcileNodeFunc(node v1.Node, monitor mntr.Monitor, pool Pool, k8s *Client) func() error {
	reconcileNode := false
	reconcileMonitor := monitor.WithField("node", node.Name)
	handleMaybe := func(maybeNode *v1.Node, maybeMonitor *mntr.Monitor) {
		if maybeNode != nil {
			reconcileNode = true
			node = *maybeNode
			reconcileMonitor = *maybeMonitor
		}
	}

	handleMaybe(reconcileLabels(node, pool, reconcileMonitor))
	handleMaybe(reconcileTaints(node, pool, reconcileMonitor))

	if !reconcileNode {
		return func() error { return nil }
	}
	return func() error {
		reconcileMonitor.Info("Reconciling node")
		return k8s.updateNode(&node)
	}
}

func reconcileTaints(node v1.Node, pool Pool, monitor mntr.Monitor) (*v1.Node, *mntr.Monitor) {
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
		return nil, nil
	}
	node.Spec.Taints = newTaints
	monitor = monitor.WithField("taints", desiredTaints)
	return &node, &monitor
}

func reconcileLabels(node v1.Node, pool Pool, monitor mntr.Monitor) (*v1.Node, *mntr.Monitor) {
	poolLabelKey := "orbos.ch/pool"
	if node.Labels[poolLabelKey] == pool.Pool {
		return nil, nil
	}
	monitor = monitor.WithField("label", fmt.Sprintf("%s=%s", poolLabelKey, pool.Pool))
	if node.Labels == nil {
		node.Labels = make(map[string]string)
	}
	node.Labels[poolLabelKey] = pool.Pool
	return &node, &monitor
}
