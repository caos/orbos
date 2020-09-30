package gce

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers"
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

		desiredKind, err := parseDesiredV0(desiredTree)
		if err != nil {
			return nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		secrets = getSecretsMap(desiredKind)
		loadBalancersSecrets, err := loadbalancers.GetSecrets(monitor, desiredKind.Loadbalancing)
		if err != nil {
			return nil, err
		}

		for k, v := range loadBalancersSecrets {
			secrets[k] = v
		}
		return secrets, nil
	}
}

func getSecretsMap(desiredKind *Desired) map[string]*secret2.Secret {
	if desiredKind.Spec.JSONKey == nil {
		desiredKind.Spec.JSONKey = &secret2.Secret{}
	}

	return map[string]*secret2.Secret{
		"jsonkey": desiredKind.Spec.JSONKey,
	}
}
