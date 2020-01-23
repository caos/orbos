package kubernetes

import (
	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/kubernetes/edge/k8s"
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
	k8sClient *k8s.Client,
	repoURL string,
	repoKey string,
	orbiterCommit string) (err error) {

	if kubeconfig != nil && kubeconfig.Value != "" {
		k8sClient.Refresh(&kubeconfig.Value)
	}

	controlplane, workers, initializeCompute, err := initialize(
		logger,
		curr,
		desired,
		nodeAgentsCurrent,
		nodeAgentsDesired,
		providerPools)
	if err != nil {
		return err
	}

	desireFirewall, ensureFirewall := firewallFuncs(desired, kubeAPIAddress.Port)
	initializeFirewall := func(_ initializedPool, computes []initializedCompute) error {
		for _, compute := range computes {
			desireFirewall(compute)
		}
		return nil
	}
	controlplane.enhance(initializeFirewall)
	controlplaneComputes, err := controlplane.computes()
	if err != nil {
		return err
	}
	workerComputes := make([]initializedCompute, 0)
	for _, workerPool := range workers {
		workerPool.enhance(initializeFirewall)
		wComputes, err := workerPool.computes()
		if err != nil {
			return err
		}
		workerComputes = append(workerComputes, wComputes...)
	}

	initializedComputes := append(controlplaneComputes, workerComputes...)

	nodeagentsDone, installNodeAgent, err := ensureNodeAgents(
		logger,
		orbiterCommit,
		repoURL,
		repoKey,
		append(controlplaneComputes, initializedComputes...),
	)
	if err != nil || !nodeagentsDone {
		logger.Debug("Node Agents are not ready yet")
		return err
	}

	if !ensureFirewall(initializedComputes) {
		logger.Debug("Firewall is not ready yet")
		return err
	}

	targetVersion := k8s.ParseString(desired.Spec.Versions.Kubernetes)
	upgradingDone, err := ensureSoftware(
		logger,
		targetVersion,
		k8sClient,
		controlplaneComputes,
		workerComputes)
	if err != nil || !upgradingDone {
		logger.Debug("Upgrading is not done yet")
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
		func(created infra.Compute, pool initializedPool) (initializedCompute, error) {
			compute := initializeCompute(created, pool)
			target := targetVersion.DefineSoftware()
			compute.desiredNodeagent.Software = &target
			return compute, installNodeAgent(compute)
		})
	if err != nil {
		logger.Debug("Scaling is not done yet")
		return err
	}

	if scalingDone {
		curr.Status = "running"
	}

	return nil
}
