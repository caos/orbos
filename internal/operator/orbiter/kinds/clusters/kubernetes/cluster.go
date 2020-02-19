package kubernetes

import (
	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/logging"
)

func ensureCluster(
	logger logging.Logger,
	desired DesiredV0,
	curr *CurrentCluster,
	nodeAgentsCurrent map[string]*common.NodeAgentCurrent,
	nodeAgentsDesired map[string]*common.NodeAgentSpec,
	providerPools map[string]map[string]infra.Pool,
	kubeAPIAddress infra.Address,
	kubeconfig *orbiter.Secret,
	psf orbiter.PushSecretsFunc,
	k8sClient *Client,
	repoURL string,
	repoKey string,
	orbiterCommit string,
	oneoff bool) (err error) {

	if kubeconfig != nil && kubeconfig.Value != "" {
		k8sClient.Refresh(&kubeconfig.Value)
	}

	controlplane, workers, initializeMachine, uninitializeMachine, err := initialize(
		logger,
		curr,
		desired,
		nodeAgentsCurrent,
		nodeAgentsDesired,
		providerPools,
		k8sClient)
	if err != nil {
		return err
	}

	desireFirewall, ensureFirewall := firewallFuncs(logger, desired, kubeAPIAddress.Port)
	initializeFirewall := func(_ initializedPool, machines []*initializedMachine) error {
		ensureFirewall(machines)
		return nil
	}
	controlplane.enhance(initializeFirewall)
	controlplaneMachines, err := controlplane.machines()
	if err != nil {
		return err
	}
	workerMachines := make([]*initializedMachine, 0)
	for _, workerPool := range workers {
		workerPool.enhance(initializeFirewall)
		wMachines, err := workerPool.machines()
		if err != nil {
			return err
		}
		workerMachines = append(workerMachines, wMachines...)
	}

	initializedMachines := append(controlplaneMachines, workerMachines...)

	nodeagentsDone, installNodeAgent, err := ensureNodeAgents(
		logger,
		orbiterCommit,
		repoURL,
		repoKey,
		initializedMachines,
	)
	if err != nil || !nodeagentsDone {
		logger.Info(false, "Node Agents are not ready yet")
		return err
	}

	if !ensureFirewall(initializedMachines) {
		logger.Info(false, "Firewall is not ready yet")
		return err
	}

	targetVersion := ParseString(desired.Spec.Versions.Kubernetes)
	upgradingDone, err := ensureSoftware(
		logger,
		targetVersion,
		k8sClient,
		controlplaneMachines,
		workerMachines)
	if err != nil || !upgradingDone {
		logger.Info(false, "Upgrading is not done yet")
		return err
	}

	var scalingDone bool
	scalingDone, err = ensureScale(
		logger,
		desired,
		kubeconfig,
		psf,
		controlplane,
		workers,
		kubeAPIAddress,
		targetVersion,
		k8sClient,
		oneoff,
		func(created infra.Machine, pool initializedPool) (initializedMachine, error) {
			machine := initializeMachine(created, pool)
			desireFirewall(*machine)
			target := targetVersion.DefineSoftware()
			machine.desiredNodeagent.Software = &target
			return *machine, installNodeAgent(*machine)
		},
		uninitializeMachine)
	if err != nil {
		return err
	}

	if scalingDone {
		curr.Status = "running"
	}
	logger.Debug("Scaling is not done yet")
	return nil
}
