package kubernetes

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/push"
	"github.com/caos/orbos/mntr"
)

func ensure(
	monitor mntr.Monitor,
	clusterID string,
	desired *DesiredV0,
	kubeAPIAddress *infra.Address,
	psf push.Func,
	k8sClient *Client,
	oneoff bool,
	controlplane initializedPool,
	controlplaneMachines []*initializedMachine,
	workers []initializedPool,
	workerMachines []*initializedMachine,
	initializeMachine initializeMachineFunc,
	uninitializeMachine uninitializeMachineFunc,
) (done bool, err error) {

	initialized := true

	for _, machine := range append(controlplaneMachines, workerMachines...) {

		if err := machine.reconcile(); err != nil {
			return false, err
		}

		machineMonitor := monitor.WithField("machine", machine.infra.ID())

		if !machine.currentMachine.FirewallIsReady {
			initialized = false
			machineMonitor.Info("Firewall is not ready yet")
		}
	}

	if !initialized {
		return false, nil
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
		return upgradingDone, err
	}

	var scalingDone bool
	scalingDone, err = ensureScale(
		monitor,
		clusterID,
		desired,
		psf,
		controlplane,
		workers,
		kubeAPIAddress,
		targetVersion,
		k8sClient,
		oneoff,
		func(created infra.Machine, pool initializedPool) initializedMachine {
			machine := initializeMachine(created, pool)
			target := targetVersion.DefineSoftware()
			machine.desiredNodeagent.Software = &target
			return *machine
		},
		uninitializeMachine)
	if !scalingDone {
		monitor.Info("Scaling is not done yet")
	}
	return scalingDone, err
}
