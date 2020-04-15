package kubernetes

import (
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/mntr"
)

func ensure(
	monitor mntr.Monitor,
	desired *DesiredV0,
	curr *CurrentCluster,
	kubeAPIAddress infra.Address,
	psf orbiter.PushSecretsFunc,
	k8sClient *Client,
	oneoff bool,
	controlplane initializedPool,
	controlplaneMachines []*initializedMachine,
	workers []initializedPool,
	workerMachines []*initializedMachine,
	initializeMachine initializeMachineFunc,
	uninitializeMachine uninitializeMachineFunc,
	installNodeAgent func(*initializedMachine) error,
) (err error) {

	initialized := true

	for _, machine := range append(controlplaneMachines, workerMachines...) {

		if err := machine.reconcile(); err != nil {
			return err
		}

		machineMonitor := monitor.WithField("machine", machine.infra.ID())
		if !machine.currentMachine.NodeAgentIsRunning {
			machineMonitor.Info("Node agent is not running on the correct version yet")
			if err := installNodeAgent(machine); err != nil {
				return err
			}

			initialized = false
		}

		if !machine.currentMachine.FirewallIsReady {
			initialized = false
			machineMonitor.Info("Firewall is not ready yet")
		}
	}

	if !initialized {
		return nil
	}

	targetVersion := ParseString(desired.Spec.Versions.Kubernetes)
	upgradingDone, err := ensureSoftware(
		monitor,
		targetVersion,
		k8sClient,
		controlplaneMachines,
		workerMachines)
	if err != nil || !upgradingDone {
		monitor.Info("Upgrading is not done yet")
		return err
	}

	var scalingDone bool
	scalingDone, err = ensureScale(
		monitor,
		desired,
		psf,
		controlplane,
		workers,
		kubeAPIAddress,
		targetVersion,
		k8sClient,
		oneoff,
		func(created infra.Machine, pool initializedPool) (initializedMachine, error) {
			machine := initializeMachine(created, pool)
			target := targetVersion.DefineSoftware()
			machine.desiredNodeagent.Software = &target
			if machine.currentMachine.NodeAgentIsRunning {
				return *machine, nil
			}
			return *machine, installNodeAgent(machine)
		},
		uninitializeMachine)
	if !scalingDone {
		monitor.Info("Scaling is not done yet")
	}
	return err
}
