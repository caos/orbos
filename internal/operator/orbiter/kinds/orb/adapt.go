package orb

import (
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/static"
	"github.com/caos/orbiter/logging"
)

func AdaptFunc(
	logger logging.Logger,
	orb *orbiter.Orb,
	orbiterCommit string,
	oneoff bool,
	deployOrbiterAndBoom bool) orbiter.AdaptFunc {
	return func(desiredTree *orbiter.Tree, secretsTree *orbiter.Tree, currentTree *orbiter.Tree) (ensureFunc orbiter.EnsureFunc, destroyFunc orbiter.DestroyFunc, secrets map[string]*orbiter.Secret, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind := &DesiredV0{Common: desiredTree.Common}
		if err := desiredTree.Original.Decode(desiredKind); err != nil {
			return nil, nil, nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredKind.Common.Version = "v0"
		desiredTree.Parsed = desiredKind

		if desiredKind.Spec.Verbose && !logger.IsVerbose() {
			logger = logger.Verbose()
		}

		secretsKind := &SecretsV0{Common: secretsTree.Common}
		if err := secretsTree.Original.Decode(secretsKind); err != nil {
			return nil, nil, nil, errors.Wrap(err, "parsing secrets failed")
		}
		secretsKind.Common.Version = "v0"
		secretsTree.Parsed = secretsKind

		providerCurrents := make(map[string]*orbiter.Tree)
		providerEnsurers := make([]orbiter.EnsureFunc, 0)
		providerDestroyers := make([]orbiter.DestroyFunc, 0)
		secrets = make(map[string]*orbiter.Secret)

		for provID, providerTree := range desiredKind.Providers {

			providerCurrent := &orbiter.Tree{}
			providerCurrents[provID] = providerCurrent

			//			providerlogger := logger.WithFields(map[string]interface{}{
			//				"provider": provID,
			//			})

			providerSecretsTree, ok := secretsKind.Providers[provID]
			if !ok {
				secretsKind.Providers[provID] = &orbiter.Tree{}
			}

			//			providerID := id + provID
			switch providerTree.Common.Kind {
			//			case "orbiter.caos.ch/GCEProvider":
			//				var lbs map[string]*infra.Ingress
			//
			//				if !kind.Spec.Destroyed && kind.Spec.ControlPlane.Provider == depID {
			//					lbs = map[string]*infra.Ingress{
			//						"kubeapi": &infra.Ingress{
			//							Pools:            []string{kind.Spec.ControlPlane.Pool},
			//							HealthChecksPath: "/healthz",
			//						},
			//					}
			//				}
			//				subassemblers[provIdx] = gce.New(providerPath, generalOverwriteSpec, gceadapter.New(providerlogger, providerID, lbs, nil, "", cfg.Params.ConnectFromOutside))
			case "orbiter.caos.ch/StaticProvider":
				//				updatesDisabled := make([]string, 0)
				//				for _, pool := range desiredKind.Spec.Workers {
				//					if pool.UpdatesDisabled {
				//						updatesDisabled = append(updatesDisabled, pool.Pool)
				//					}
				//				}
				//
				//				if desiredKind.Spec.ControlPlane.UpdatesDisabled {
				//					updatesDisabled = append(updatesDisabled, desiredKind.Spec.ControlPlane.Pool)
				//				}

				providerEnsurer, providerDestroyer, providerSecrets, err := static.AdaptFunc(logger, orb.Masterkey, provID)(providerTree, providerSecretsTree, providerCurrent)
				if err != nil {
					return nil, nil, nil, err
				}
				providerEnsurers = append(providerEnsurers, providerEnsurer)
				providerDestroyers = append(providerDestroyers, providerDestroyer)
				for path, secret := range providerSecrets {
					secrets[orbiter.JoinPath(provID, path)] = secret
				}
			default:
				return nil, nil, nil, errors.Errorf("unknown provider kind %s", providerTree.Common.Kind)
			}
		}

		var provCurr map[string]interface{}
		ensureProviders := func(psf orbiter.PushSecretsFunc, nodeAgentsCurrent map[string]*common.NodeAgentCurrent, nodeAgentsDesired map[string]*common.NodeAgentSpec) (map[string]interface{}, error) {

			if provCurr != nil {
				return provCurr, nil
			}

			provCurr = make(map[string]interface{})
			for _, ensurer := range providerEnsurers {
				if err := ensurer(psf, nodeAgentsCurrent, nodeAgentsDesired); err != nil {
					return nil, err
				}
			}

			for currKey, currVal := range providerCurrents {
				provCurr[currKey] = currVal.Parsed
			}
			return provCurr, nil
		}
		destroyProviders := func() (map[string]interface{}, error) {
			if provCurr != nil {
				return provCurr, nil
			}

			provCurr = make(map[string]interface{})
			for _, destroyer := range providerDestroyers {
				if err := destroyer(); err != nil {
					return nil, err
				}
			}

			for currKey, currVal := range providerCurrents {
				provCurr[currKey] = currVal.Parsed
			}
			return provCurr, nil
		}

		clusterCurrents := make(map[string]*orbiter.Tree)
		clusterEnsurers := make([]orbiter.EnsureFunc, 0)
		clusterDestroyers := make([]orbiter.DestroyFunc, 0)
		for clusterID, clusterTree := range desiredKind.Clusters {

			clusterCurrent := &orbiter.Tree{}
			clusterCurrents[clusterID] = clusterCurrent

			clusterSecretsTree, ok := secretsKind.Clusters[clusterID]
			if !ok {
				return nil, nil, nil, errors.Errorf("no secrets found for cluster %s", clusterID)
			}

			switch clusterTree.Common.Kind {
			case "orbiter.caos.ch/KubernetesCluster":
				clusterEnsurer, clusterDestroyer, clusterSecrets, err := kubernetes.AdaptFunc(logger, orb, orbiterCommit, clusterID, oneoff, deployOrbiterAndBoom, ensureProviders, destroyProviders)(clusterTree, clusterSecretsTree, clusterCurrent)
				if err != nil {
					return nil, nil, nil, err
				}
				clusterEnsurers = append(clusterEnsurers, clusterEnsurer)
				clusterDestroyers = append(clusterDestroyers, clusterDestroyer)
				for path, secret := range clusterSecrets {
					secrets[orbiter.JoinPath(clusterID, path)] = secret
				}

				//				subassemblers[provIdx] = static.New(providerPath, generalOverwriteSpec, staticadapter.New(providerlogger, providerID, "/healthz", updatesDisabled, cfg.NodeAgent))
			default:
				return nil, nil, nil, errors.Errorf("unknown cluster kind %s", clusterTree.Common.Kind)
			}
		}

		currentTree.Parsed = &Current{
			Common: &orbiter.Common{
				Kind:    "orbiter.caos.ch/Orb",
				Version: "v0",
			},
			Clusters:  clusterCurrents,
			Providers: providerCurrents,
		}

		return func(psf orbiter.PushSecretsFunc, nodeAgentsCurrent map[string]*common.NodeAgentCurrent, nodeAgentsDesired map[string]*common.NodeAgentSpec) (err error) {
				defer func() {
					err = errors.Wrapf(err, "ensuring %s failed", desiredKind.Common.Kind)
				}()

				for _, ensurer := range clusterEnsurers {
					if err := ensurer(psf, nodeAgentsCurrent, nodeAgentsDesired); err != nil {
						return err
					}
				}
				return nil
			}, func() error {
				defer func() {
					err = errors.Wrapf(err, "ensuring %s failed", desiredKind.Common.Kind)
				}()

				for _, destroyer := range clusterDestroyers {
					if err := destroyer(); err != nil {
						return err
					}
				}
				return nil
			}, secrets, nil
	}
}
