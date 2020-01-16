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
	repoURL string,
	repoKey string,
	masterKey string,
	orbiterCommit string,
	destroy bool) orbiter.AdaptFunc {
	return func(desiredTree *orbiter.Tree, secretsTree *orbiter.Tree, currentTree *orbiter.Tree) (ensureFunc orbiter.EnsureFunc, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind := &DesiredV0{Common: desiredTree.Common}
		if err := desiredTree.Original.Decode(desiredKind); err != nil {
			return nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredKind.Common.Version = "v0"
		desiredTree.Parsed = desiredKind

		secretsKind := &SecretsV0{Common: secretsTree.Common}
		if err := secretsTree.Original.Decode(secretsKind); err != nil {
			return nil, errors.Wrap(err, "parsing secrets failed")
		}
		secretsKind.Common.Version = "v0"
		secretsTree.Parsed = secretsKind

		clusterCurrents := make(map[string]*orbiter.Tree)
		clusterEnsurers := make([]orbiter.EnsureFunc, 0)
		for clusterID, clusterTree := range desiredKind.Deps {

			clusterCurrent := &orbiter.Tree{}
			clusterCurrents[clusterID] = clusterCurrent

			clusterSecretsTree, ok := secretsKind.Deps[clusterID]
			if !ok {
				return nil, errors.Errorf("no secrets found for cluster %s", clusterID)
			}

			switch clusterTree.Common.Kind {
			case "orbiter.caos.ch/KubernetesCluster":
				clusterEnsurer, err := kubernetes.AdaptFunc(logger, repoURL, repoKey, masterKey, orbiterCommit, clusterID, destroy)(clusterTree, clusterSecretsTree, clusterCurrent)
				if err != nil {
					return nil, err
				}
				clusterEnsurers = append(clusterEnsurers, clusterEnsurer)

				//				subassemblers[provIdx] = static.New(providerPath, generalOverwriteSpec, staticadapter.New(providerlogger, providerID, "/healthz", updatesDisabled, cfg.NodeAgent))
			default:
				return nil, errors.Errorf("unknown cluster kind %s", clusterTree.Common.Kind)
			}
		}

		currentTree.Parsed = &Current{
			Common: &orbiter.Common{
				Kind:    "orbiter.caos.ch/KubernetesCluster",
				Version: "v0",
			},
			Deps: clusterCurrents,
		}

		return func(nodeAgentsCurrent map[string]*common.NodeAgentCurrent, nodeAgentsDesired map[string]*common.NodeAgentSpec) (err error) {
			defer func() {
				err = errors.Wrapf(err, "ensuring %s failed", desiredKind.Common.Kind)
			}()
			for _, ensurer := range clusterEnsurers {
				if err := ensurer(nodeAgentsCurrent, nodeAgentsDesired); err != nil {
					return err
				}
			}
			return nil
		}, nil
	}
}
