package orb

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers"
	"github.com/caos/orbos/internal/orb"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"

	"github.com/caos/orbos/mntr"
)

func SecretsFunc(
	orb *orb.Orb) secret.Func {
	return func(monitor mntr.Monitor, desiredTree *tree.Tree) (secrets map[string]*secret.Secret, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind, err := parseDesiredV0(desiredTree)
		if err != nil {
			return nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		secrets = make(map[string]*secret.Secret)

		for provID, providerTree := range desiredKind.Providers {

			providerSecrets, err := providers.GetSecrets(
				monitor.WithFields(map[string]interface{}{"provider": provID}),
				orb.Masterkey,
				providerTree,
			)
			if err != nil {
				return nil, err
			}

			for path, providerSecret := range providerSecrets {
				secrets[secret.JoinPath(provID, path)] = providerSecret
			}
		}

		for clusterID, clusterTree := range desiredKind.Clusters {

			clusterSecrets, err := clusters.GetSecrets(
				monitor.WithFields(map[string]interface{}{"cluster": clusterID}),
				orb,
				clusterTree,
			)
			if err != nil {
				return nil, err
			}

			for path, clusterSecret := range clusterSecrets {
				secrets[secret.JoinPath(clusterID, path)] = clusterSecret
			}
		}

		return secrets, nil
	}
}
