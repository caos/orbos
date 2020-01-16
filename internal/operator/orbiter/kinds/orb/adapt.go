package orb

import (
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbiter/logging"
)

func AdaptFunc(
	logger logging.Logger,
	orb *orbiter.Orb,
	orbiterCommit string,
	destroy bool,
	oneoff bool) orbiter.AdaptFunc {
	return func(desiredTree *orbiter.Tree, secretsTree *orbiter.Tree, currentTree *orbiter.Tree) (ensureFunc orbiter.EnsureFunc, readSecretFunc orbiter.ReadSecretFunc, writeSecretFunc orbiter.WriteSecretFunc, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind := &DesiredV0{Common: desiredTree.Common}
		if err := desiredTree.Original.Decode(desiredKind); err != nil {
			return nil, nil, nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredKind.Common.Version = "v0"
		desiredTree.Parsed = desiredKind

		secretsKind := &SecretsV0{Common: secretsTree.Common}
		if err := secretsTree.Original.Decode(secretsKind); err != nil {
			return nil, nil, nil, errors.Wrap(err, "parsing secrets failed")
		}
		secretsKind.Common.Version = "v0"
		secretsTree.Parsed = secretsKind

		clusterCurrents := make(map[string]*orbiter.Tree)
		clusterEnsurers := make([]orbiter.EnsureFunc, 0)
		clusterSecretReaders := make(map[string]orbiter.ReadSecretFunc)
		clusterSecretWriters := make(map[string]orbiter.WriteSecretFunc)
		for clusterID, clusterTree := range desiredKind.Deps {

			clusterCurrent := &orbiter.Tree{}
			clusterCurrents[clusterID] = clusterCurrent

			clusterSecretsTree, ok := secretsKind.Deps[clusterID]
			if !ok {
				return nil, nil, nil, errors.Errorf("no secrets found for cluster %s", clusterID)
			}

			switch clusterTree.Common.Kind {
			case "orbiter.caos.ch/KubernetesCluster":
				clusterEnsurer, clusterSecretReader, clusterSecretWriter, err := kubernetes.AdaptFunc(logger, orb, orbiterCommit, clusterID, destroy, oneoff)(clusterTree, clusterSecretsTree, clusterCurrent)
				if err != nil {
					return nil, nil, nil, err
				}
				clusterEnsurers = append(clusterEnsurers, clusterEnsurer)
				clusterSecretReaders[clusterID] = clusterSecretReader
				clusterSecretWriters[clusterID] = clusterSecretWriter

				//				subassemblers[provIdx] = static.New(providerPath, generalOverwriteSpec, staticadapter.New(providerlogger, providerID, "/healthz", updatesDisabled, cfg.NodeAgent))
			default:
				return nil, nil, nil, errors.Errorf("unknown cluster kind %s", clusterTree.Common.Kind)
			}
		}

		currentTree.Parsed = &Current{
			Common: &orbiter.Common{
				Kind:    "orbiter.caos.ch/KubernetesCluster",
				Version: "v0",
			},
			Deps: clusterCurrents,
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
			}, func(path []string) (string, error) {
				return orbiter.AdaptReadSecret(path, clusterSecretReaders, nil)
			}, func(path []string, value string) error {
				return orbiter.AdaptWriteSecret(path, value, clusterSecretWriters, nil)
			}, nil
	}
}
