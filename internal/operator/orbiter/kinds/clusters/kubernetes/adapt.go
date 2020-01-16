package kubernetes

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/static"
	"github.com/caos/orbiter/logging"
)

func AdaptFunc(
	logger logging.Logger,
	orb *orbiter.Orb,
	orbiterCommit string,
	id string,
	destroy bool,
	takeoff bool) orbiter.AdaptFunc {
	return func(desiredTree *orbiter.Tree, secretsTree *orbiter.Tree, currentTree *orbiter.Tree) (ensureFunc orbiter.EnsureFunc, readSecretFunc orbiter.ReadSecretFunc, writeSecretFunc orbiter.WriteSecretFunc, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind := &DesiredV0{Common: *desiredTree.Common}
		if err := desiredTree.Original.Decode(desiredKind); err != nil {
			return nil, nil, nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredKind.Common.Version = "v0"
		desiredTree.Parsed = desiredKind

		secretsKind := &SecretsV0{
			Common:  *secretsTree.Common,
			Secrets: Secrets{Kubeconfig: &orbiter.Secret{Masterkey: orb.Masterkey}},
		}
		if err := secretsTree.Original.Decode(secretsKind); err != nil {
			return nil, nil, nil, errors.Wrap(err, "parsing secrets failed")
		}
		secretsKind.Common.Version = "v0"
		secretsTree.Parsed = secretsKind

		if secretsKind.Secrets.Kubeconfig != nil && secretsKind.Secrets.Kubeconfig.Value != "" {
			if err := ensureArtifacts(logger, secretsKind.Secrets.Kubeconfig, orb, takeoff, desiredKind.Spec.Versions.Orbiter, desiredKind.Spec.Versions.Boom); err != nil {
				return nil, nil, nil, err
			}
		}

		providerCurrents := make(map[string]*orbiter.Tree)
		providerEnsurers := make([]orbiter.EnsureFunc, 0)
		providerSecretReaders := make(map[string]orbiter.ReadSecretFunc)
		providerSecretWriters := make(map[string]orbiter.WriteSecretFunc)
		for provID, providerTree := range desiredKind.Deps {

			providerCurrent := &orbiter.Tree{}
			providerCurrents[provID] = providerCurrent

			//			providerlogger := logger.WithFields(map[string]interface{}{
			//				"provider": provID,
			//			})

			providerSecretsTree, ok := secretsKind.Deps[provID]
			if !ok {
				secretsKind.Deps[provID] = &orbiter.Tree{}
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
				updatesDisabled := make([]string, 0)
				for _, pool := range desiredKind.Spec.Workers {
					if pool.UpdatesDisabled {
						updatesDisabled = append(updatesDisabled, pool.Pool)
					}
				}

				if desiredKind.Spec.ControlPlane.UpdatesDisabled {
					updatesDisabled = append(updatesDisabled, desiredKind.Spec.ControlPlane.Pool)
				}

				providerEnsurer, providerSecretReader, providerSecretWriter, err := static.AdaptFunc(logger, orb.Masterkey, fmt.Sprintf("%s:%s", id, provID))(providerTree, providerSecretsTree, providerCurrent)
				if err != nil {
					return nil, nil, nil, err
				}
				providerEnsurers = append(providerEnsurers, providerEnsurer)
				providerSecretReaders[provID] = providerSecretReader
				providerSecretWriters[provID] = providerSecretWriter
				//				subassemblers[provIdx] = static.New(providerPath, generalOverwriteSpec, staticadapter.New(providerlogger, providerID, "/healthz", updatesDisabled, cfg.NodeAgent))
			default:
				return nil, nil, nil, errors.Errorf("unknown provider kind %s", providerTree.Common.Kind)
			}
		}

		current := &CurrentCluster{}
		currentTree.Parsed = &Current{
			Common: orbiter.Common{
				Kind:    "orbiter.caos.ch/KubernetesCluster",
				Version: "v0",
			},
			Deps:    providerCurrents,
			Current: *current,
		}

		secretsMap := map[string]*orbiter.Secret{
			"kubeconfig": secretsKind.Secrets.Kubeconfig,
		}

		return func(psf orbiter.PushSecretsFunc, nodeAgentsCurrent map[string]*common.NodeAgentCurrent, nodeAgentsDesired map[string]*common.NodeAgentSpec) (err error) {
				defer func() {
					err = errors.Wrapf(err, "ensuring %s failed", desiredKind.Common.Kind)
				}()
				for _, ensurer := range providerEnsurers {
					if err := ensurer(psf, nodeAgentsCurrent, nodeAgentsDesired); err != nil {
						return err
					}
				}

				providers := make(map[string]interface{})
				for provID, providerCurrent := range providerCurrents {
					providers[provID] = providerCurrent.Parsed
				}

				return ensure(
					logger,
					*desiredKind,
					current,
					providers,
					nodeAgentsCurrent,
					nodeAgentsDesired,
					psf,
					secretsKind.Secrets.Kubeconfig,
					orb.URL,
					orb.Repokey,
					orbiterCommit,
					destroy)
			}, func(path []string) (string, error) {
				return orbiter.AdaptReadSecret(path, providerSecretReaders, secretsMap)
			}, func(path []string, value string) error {
				return orbiter.AdaptWriteSecret(path, value, providerSecretWriters, secretsMap)
			}, nil
	}
}
