package orb

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers"
	secret2 "github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"

	"github.com/caos/orbos/mntr"
)

func SecretsFunc() secret2.Func {
	return func(monitor mntr.Monitor, desiredTree *tree.Tree) (secrets map[string]*secret2.Secret, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind, err := ParseDesiredV0(desiredTree)
		if err != nil {
			return nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		secrets = make(map[string]*secret2.Secret)

		for provID, providerTree := range desiredKind.Providers {

			providerSecrets, err := providers.GetSecrets(
				monitor.WithFields(map[string]interface{}{"provider": provID}),
				providerTree,
			)
			if err != nil {
				return nil, err
			}

			for path, providerSecret := range providerSecrets {
				secrets[secret2.JoinPath(provID, path)] = providerSecret
			}
		}

		for clusterID, clusterTree := range desiredKind.Clusters {

			clusterSecrets, err := clusters.GetSecrets(
				monitor.WithFields(map[string]interface{}{"cluster": clusterID}),
				clusterTree,
			)
			if err != nil {
				return nil, err
			}

			for path, clusterSecret := range clusterSecrets {
				secrets[secret2.JoinPath(clusterID, path)] = clusterSecret
			}
		}

		return secrets, nil
	}
}
