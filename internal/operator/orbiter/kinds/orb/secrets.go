package orb

import (
	"github.com/caos/orbiter/internal/orb"
	"github.com/caos/orbiter/internal/secret"
	"github.com/caos/orbiter/internal/tree"
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/static"
	"github.com/caos/orbiter/mntr"
)

func SecretsFunc(
	orb *orb.Orb) secret.Func {
	return func(monitor mntr.Monitor, desiredTree *tree.Tree, currentTree *tree.Tree) (secrets map[string]*secret.Secret, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind := &DesiredV0{Common: desiredTree.Common}
		if err := desiredTree.Original.Decode(desiredKind); err != nil {
			return nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredKind.Common.Version = "v0"
		desiredTree.Parsed = desiredKind

		if err := desiredKind.validate(); err != nil {
			return nil, err
		}

		if desiredKind.Spec.Verbose && !monitor.IsVerbose() {
			monitor = monitor.Verbose()
		}

		providerCurrents := make(map[string]*tree.Tree)
		secrets = make(map[string]*secret.Secret)

		for provID, providerTree := range desiredKind.Providers {

			providerCurrent := &tree.Tree{}
			providerCurrents[provID] = providerCurrent

			switch providerTree.Common.Kind {
			case "orbiter.caos.ch/StaticProvider":
				providerSecrets, err := static.SecretsFunc(
					orb.Masterkey,
				)(
					monitor.WithFields(map[string]interface{}{"provider": provID}),
					providerTree,
					providerCurrent)

				if err != nil {
					return nil, err
				}

				for path, providerSecret := range providerSecrets {
					secrets[secret.JoinPath(provID, path)] = providerSecret
				}
			default:
				return nil, errors.Errorf("unknown provider kind %s", providerTree.Common.Kind)
			}
		}

		clusterCurrents := make(map[string]*tree.Tree)
		for clusterID, clusterTree := range desiredKind.Clusters {

			clusterCurrent := &tree.Tree{}
			clusterCurrents[clusterID] = clusterCurrent

			switch clusterTree.Common.Kind {
			case "orbiter.caos.ch/KubernetesCluster":
				clusterSecrets, err := kubernetes.SecretFunc(orb)(
					monitor.WithFields(map[string]interface{}{"cluster": clusterID}),
					clusterTree,
					clusterCurrent)

				if err != nil {
					return nil, err
				}

				for path, clusterSecret := range clusterSecrets {
					secrets[secret.JoinPath(clusterID, path)] = clusterSecret
				}

			default:
				return nil, errors.Errorf("unknown cluster kind %s", clusterTree.Common.Kind)
			}
		}

		currentTree.Parsed = &Current{
			Common: &tree.Common{
				Kind:    "orbiter.caos.ch/Orb",
				Version: "v0",
			},
			Clusters:  clusterCurrents,
			Providers: providerCurrents,
		}

		return secrets, nil
	}
}
