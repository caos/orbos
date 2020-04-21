package orb

import (
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/static"
	"github.com/caos/orbiter/mntr"
)

func AdaptFunc(
	orb *orbiter.Orb,
	orbiterCommit string,
	oneoff bool,
	deployOrbiterAndBoom bool) orbiter.AdaptFunc {
	return func(monitor mntr.Monitor, desiredTree *orbiter.Tree, currentTree *orbiter.Tree) (queryFunc orbiter.QueryFunc, destroyFunc orbiter.DestroyFunc, secrets map[string]*orbiter.Secret, migrate bool, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind := &DesiredV0{Common: desiredTree.Common}
		if err := desiredTree.Original.Decode(desiredKind); err != nil {
			return nil, nil, nil, migrate, errors.Wrap(err, "parsing desired state failed")
		}
		desiredKind.Common.Version = "v0"
		desiredTree.Parsed = desiredKind

		if err := desiredKind.validate(); err != nil {
			return nil, nil, nil, migrate, err
		}

		if desiredKind.Spec.Verbose && !monitor.IsVerbose() {
			monitor = monitor.Verbose()
		}

		providerCurrents := make(map[string]*orbiter.Tree)
		providerQueriers := make([]orbiter.QueryFunc, 0)
		providerDestroyers := make([]orbiter.DestroyFunc, 0)
		secrets = make(map[string]*orbiter.Secret)

		whitelistChan := make(chan []*orbiter.CIDR)

		for provID, providerTree := range desiredKind.Providers {

			providerCurrent := &orbiter.Tree{}
			providerCurrents[provID] = providerCurrent

			//			providermonitor := monitor.WithFields(map[string]interface{}{
			//				"provider": provID,
			//			})

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
			//				subassemblers[provIdx] = gce.New(providerPath, generalOverwriteSpec, gceadapter.New(providermonitor, providerID, lbs, nil, "", cfg.Params.ConnectFromOutside))
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

				providerQuerier, providerDestroyer, providerSecrets, pMigrate, err := static.AdaptFunc(
					orb.Masterkey,
					provID,
					func() []*orbiter.CIDR {
						monitor.Debug("Reading whitelist")
						return <-whitelistChan
					},
				)(
					monitor.WithFields(map[string]interface{}{"provider": provID}),
					providerTree,
					providerCurrent)
				if err != nil {
					return nil, nil, nil, migrate, err
				}
				if pMigrate {
					migrate = true
				}
				providerQueriers = append(providerQueriers, providerQuerier)
				providerDestroyers = append(providerDestroyers, providerDestroyer)
				for path, secret := range providerSecrets {
					secrets[orbiter.JoinPath(provID, path)] = secret
				}
			default:
				return nil, nil, nil, migrate, errors.Errorf("unknown provider kind %s", providerTree.Common.Kind)
			}
		}

		var provCurr map[string]interface{}
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
		clusterQueriers := make([]orbiter.QueryFunc, 0)
		clusterDestroyers := make([]orbiter.DestroyFunc, 0)
		for clusterID, clusterTree := range desiredKind.Clusters {

			clusterCurrent := &orbiter.Tree{}
			clusterCurrents[clusterID] = clusterCurrent

			switch clusterTree.Common.Kind {
			case "orbiter.caos.ch/KubernetesCluster":
				clusterQuerier, clusterDestroyer, clusterSecrets, cMigrate, err := kubernetes.AdaptFunc(
					orb,
					orbiterCommit,
					clusterID,
					oneoff,
					deployOrbiterAndBoom,
					destroyProviders,
					func(whitelist []*orbiter.CIDR) {
						go func() {
							monitor.Debug("Sending whitelist")
							whitelistChan <- whitelist
							close(whitelistChan)
						}()
						monitor.Debug("Whitelist sent")
					},
				)(
					monitor.WithFields(map[string]interface{}{"cluster": clusterID}),
					clusterTree,
					clusterCurrent)
				if err != nil {
					return nil, nil, nil, migrate, err
				}
				if cMigrate {
					migrate = true
				}
				clusterQueriers = append(clusterQueriers, clusterQuerier)
				clusterDestroyers = append(clusterDestroyers, clusterDestroyer)
				for path, secret := range clusterSecrets {
					secrets[orbiter.JoinPath(clusterID, path)] = secret
				}

				//				subassemblers[provIdx] = static.New(providerPath, generalOverwriteSpec, staticadapter.New(providermonitor, providerID, "/healthz", updatesDisabled, cfg.NodeAgent))
			default:
				return nil, nil, nil, migrate, errors.Errorf("unknown cluster kind %s", clusterTree.Common.Kind)
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

		return func(nodeAgentsCurrent map[string]*common.NodeAgentCurrent, nodeAgentsDesired map[string]*common.NodeAgentSpec, _ map[string]interface{}) (ensureFunc orbiter.EnsureFunc, err error) {

				providerEnsurers := make([]orbiter.EnsureFunc, 0)
				queriedProviders := make(map[string]interface{})
				for _, querier := range providerQueriers {
					ensurer, err := querier(nodeAgentsCurrent, nodeAgentsDesired, nil)
					if err != nil {
						return nil, err
					}
					providerEnsurers = append(providerEnsurers, ensurer)
				}

				for currKey, currVal := range providerCurrents {
					queriedProviders[currKey] = currVal.Parsed
				}

				clusterEnsurers := make([]orbiter.EnsureFunc, 0)
				for _, querier := range clusterQueriers {
					ensurer, err := querier(nodeAgentsCurrent, nodeAgentsDesired, queriedProviders)
					if err != nil {
						return nil, err
					}
					clusterEnsurers = append(clusterEnsurers, ensurer)
				}

				return func(psf orbiter.PushSecretsFunc) (err error) {
					defer func() {
						err = errors.Wrapf(err, "ensuring %s failed", desiredKind.Common.Kind)
					}()

					for _, ensurer := range append(providerEnsurers, clusterEnsurers...) {
						if err := ensurer(psf); err != nil {
							return err
						}
					}

					return nil
				}, nil
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
			}, secrets, migrate, nil
	}
}
