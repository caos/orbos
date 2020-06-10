package kubernetes

import (
	"github.com/caos/orbos/internal/api"
	"github.com/pkg/errors"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/mntr"
)

func query(
	monitor mntr.Monitor,
	clusterID string,
	desired *DesiredV0,
	current *CurrentCluster,
	providerCurrents map[string]interface{},
	nodeAgentsCurrent map[string]*common.NodeAgentCurrent,
	nodeAgentsDesired map[string]*common.NodeAgentSpec,
	k8sClient *Client,
	repoURL string,
	repoKey string,
	orbiterCommit string,
	oneoff bool) (orbiter.EnsureFunc, error) {

	current.Machines = make(map[string]*Machine)

	cloudPools := make(map[string]map[string]infra.Pool)
	var kubeAPIAddress infra.Address

	for providerName, provider := range providerCurrents {
		if cloudPools[providerName] == nil {
			cloudPools[providerName] = make(map[string]infra.Pool)
		}
		prov := provider.(infra.ProviderCurrent)
		providerPools := prov.Pools()
		providerIngresses := prov.Ingresses()
		for providerPoolName, providerPool := range providerPools {
			cloudPools[providerName][providerPoolName] = providerPool
			if desired.Spec.ControlPlane.Provider == providerName && desired.Spec.ControlPlane.Pool == providerPoolName {
				kubeAPIAddress = providerIngresses["kubeapi"]
				monitor.WithFields(map[string]interface{}{
					"address": kubeAPIAddress,
				}).Debug("Found kubernetes api address")
			}
		}
	}

	if err := poolIsConfigured(&desired.Spec.ControlPlane, cloudPools); err != nil {
		return nil, err
	}

	for _, w := range desired.Spec.Workers {
		if err := poolIsConfigured(w, cloudPools); err != nil {
			return nil, err
		}
	}

	queryNodeAgent, installNodeAgent := nodeAgentFuncs(
		monitor,
		orbiterCommit,
		repoURL,
		repoKey,
	)

	controlplane, controlplaneMachines, workers, workerMachines, initializeMachine, uninitializeMachine, err := initialize(
		monitor,
		current,
		*desired,
		nodeAgentsCurrent,
		nodeAgentsDesired,
		cloudPools,
		k8sClient,
		func(machine *initializedMachine) {
			queryNodeAgent(machine)
			firewallFunc(monitor, *desired, kubeAPIAddress.Port)(machine)
		})

	return func(psf api.SecretFunc) error {
		return ensure(
			monitor,
			clusterID,
			desired,
			current,
			kubeAPIAddress,
			psf,
			k8sClient,
			oneoff,
			controlplane,
			controlplaneMachines,
			workers,
			workerMachines,
			initializeMachine,
			uninitializeMachine,
			installNodeAgent)
	}, err
}

func poolIsConfigured(poolSpec *Pool, infra map[string]map[string]infra.Pool) error {
	prov, ok := infra[poolSpec.Provider]
	if !ok {
		return errors.Errorf("provider %s not configured", poolSpec.Provider)
	}
	if _, ok := prov[poolSpec.Pool]; !ok {
		return errors.Errorf("pool %s not configured on provider %s", poolSpec.Provider, poolSpec.Pool)
	}
	return nil
}
