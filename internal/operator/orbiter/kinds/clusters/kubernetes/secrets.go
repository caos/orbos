package kubernetes

import (
	secret2 "github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"

	"github.com/caos/orbos/mntr"
)

func SecretFunc() secret2.Func {

	return func(monitor mntr.Monitor, desiredTree *tree.Tree) (secrets map[string]*secret2.Secret, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind, err := parseDesiredV0(desiredTree)
		if err != nil {
			return nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		return getSecretsMap(desiredKind), nil
	}
}

func getSecretsMap(desiredKind *DesiredV0) map[string]*secret2.Secret {
	if desiredKind.Spec.Kubeconfig == nil {
		desiredKind.Spec.Kubeconfig = &secret2.Secret{}
	}
	return map[string]*secret2.Secret{
		"kubeconfig": desiredKind.Spec.Kubeconfig,
	}
}
