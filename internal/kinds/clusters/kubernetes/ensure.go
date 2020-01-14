package kubernetes

import (
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/core/operator/orbiter"
)

func ensure(
	desired *DesiredV0,
	current *CurrentCluster,
	providerCurrents map[string]interface{},
	nodeAgentsCurrent map[string]*orbiter.NodeAgentCurrent,
	nodeAgentsDesired map[string]*orbiter.NodeAgentSpec) error {

	poolIsConfigured := func(poolSpec *Pool, infra map[string]map[string]infra.Pool) error {
		prov, ok := infra[poolSpec.Provider]
		if !ok {
			return errors.Errorf("provider %s not configured", poolSpec.Provider)
		}
		if _, ok := prov[poolSpec.Pool]; !ok {
			return errors.Errorf("pool %s not configured on provider %s", poolSpec.Provider, poolSpec.Pool)
		}
		return nil
	}

	current.Status = "maintaining"
	current.Computes = make(map[string]*Compute)

	cloudPools := make(map[string]map[string]infra.Pool)
	providersCleanupped := make([]<-chan error, 0)
	var kubeAPIAddress infra.Address

	for providerName, provider := range ensuredDependencies {
		if cloudPools[providerName] == nil {
			cloudPools[providerName] = make(map[string]infra.Pool)
		}
		prov := provider.(infra.ProviderCurrent)
		providerPools := prov.Pools()
		providerIngresses := prov.Ingresses()
		providerCleanupped := prov.Cleanupped()
		providersCleanupped = append(providersCleanupped, providerCleanupped)
		for providerPoolName, providerPool := range providerPools {
			cloudPools[providerName][providerPoolName] = providerPool
			if spec.ControlPlane.Provider == providerName && spec.ControlPlane.Pool == providerPoolName {
				kubeAPIAddress = providerIngresses["kubeapi"]
				cfg.Params.Logger.WithFields(map[string]interface{}{
					"address": kubeAPIAddress,
				}).Debug("Found kubernetes api address")
			}
		}
	}

	if err := poolIsConfigured(spec.ControlPlane, cloudPools); err != nil {
		return err
	}

	for _, w := range spec.Workers {
		if err := poolIsConfigured(w, cloudPools); err != nil {
			return err
		}
	}

	k8sClient := k8s.New(cfg.Params.Logger, nil)
	if err := ensureCluster(&cfg, curr, cloudPools, kubeAPIAddress, secrets, k8sClient); err != nil {
		return errors.Wrap(err, "ensuring cluster failed")
	}

	if spec.Destroyed {
		return infra.Destroy(ensuredDependencies)
	}

	for _, cleanupped := range providersCleanupped {
		if err := <-cleanupped; err != nil {
			return err
		}
	}

	return nil
}
