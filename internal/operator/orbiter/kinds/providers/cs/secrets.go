package cs

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"

	"github.com/caos/orbos/mntr"
)

func SecretsFunc() secret.Func {
	return func(monitor mntr.Monitor, desiredTree *tree.Tree) (secrets map[string]*secret.Secret, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind, err := parseDesired(desiredTree)
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

func getSecretsMap(desiredKind *Desired) map[string]*secret.Secret {
	if desiredKind.Spec.APIToken == nil {
		desiredKind.Spec.APIToken = &secret.Secret{}
	}

	if desiredKind.Spec.SSHKey == nil {
		desiredKind.Spec.SSHKey = &SSHKey{}
	}

	if desiredKind.Spec.SSHKey.Public == nil {
		desiredKind.Spec.SSHKey.Public = &secret.Secret{}
	}

	if desiredKind.Spec.SSHKey.Private == nil {
		desiredKind.Spec.SSHKey.Private = &secret.Secret{}
	}

	return map[string]*secret.Secret{
		"jsonkey":       desiredKind.Spec.APIToken,
		"sshkeyprivate": desiredKind.Spec.SSHKey.Private,
		"sshkeypublic":  desiredKind.Spec.SSHKey.Public,
	}
}
