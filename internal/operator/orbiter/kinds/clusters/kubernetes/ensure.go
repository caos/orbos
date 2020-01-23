package kubernetes

import (
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/kubernetes/edge/k8s"
	"github.com/caos/orbiter/logging"
)

func ensure(
	logger logging.Logger,
	desired DesiredV0,
	current *CurrentCluster,
	providerCurrents map[string]interface{},
	nodeAgentsCurrent map[string]*common.NodeAgentCurrent,
	nodeAgentsDesired map[string]*common.NodeAgentSpec,
	psf orbiter.PushSecretsFunc,
	kubeconfig *orbiter.Secret,
	repoURL string,
	repoKey string,
	orbiterCommit string,
	oneoff bool) error {

	current.Status = "maintaining"
	current.Computes = make(map[string]*Compute)

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
				logger.WithFields(map[string]interface{}{
					"address": kubeAPIAddress,
				}).Debug("Found kubernetes api address")
			}
		}
	}

	if err := poolIsConfigured(&desired.Spec.ControlPlane, cloudPools); err != nil {
		return err
	}

	for _, w := range desired.Spec.Workers {
		if err := poolIsConfigured(w, cloudPools); err != nil {
			return err
		}
	}

	k8sClient := k8s.New(logger, nil)
	if err := ensureCluster(
		logger,
		desired,
		current,
		nodeAgentsCurrent,
		nodeAgentsDesired,
		cloudPools,
		kubeAPIAddress,
		kubeconfig,
		psf,
		k8sClient,
		repoURL,
		repoKey,
		orbiterCommit,
		oneoff); err != nil {
		return errors.Wrap(err, "ensuring cluster failed")
	}

	return nil
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
