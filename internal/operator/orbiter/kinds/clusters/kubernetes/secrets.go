package kubernetes

import (
	"github.com/caos/orbiter/internal/orb"
	"github.com/caos/orbiter/internal/secret"
	"github.com/caos/orbiter/internal/tree"
	"github.com/pkg/errors"

	"github.com/caos/orbiter/mntr"
)

func SecretFunc(orb *orb.Orb) secret.Func {

	return func(monitor mntr.Monitor, desiredTree *tree.Tree) (secrets map[string]*secret.Secret, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind, err := parseDesiredV0(desiredTree, orb.Masterkey)
		if err != nil {
			return nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		initializeNecessarySecrets(desiredKind, orb.Masterkey)

		return getSecretsMap(desiredKind), nil
	}
}

func getSecretsMap(desiredKind *DesiredV0) map[string]*secret.Secret {
	return map[string]*secret.Secret{
		"kubeconfig": desiredKind.Spec.Kubeconfig,
	}
}