package static

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

func getSecretsMap(desiredKind *DesiredV0) map[string]*secret.Secret {

	if desiredKind.Spec.Keys.BootstrapKeyPrivate == nil {
		desiredKind.Spec.Keys.BootstrapKeyPrivate = &secret.Secret{}
	}

	if desiredKind.Spec.Keys.BootstrapKeyPublic == nil {
		desiredKind.Spec.Keys.BootstrapKeyPublic = &secret.Secret{}
	}

	if desiredKind.Spec.Keys.MaintenanceKeyPrivate == nil {
		desiredKind.Spec.Keys.MaintenanceKeyPrivate = &secret.Secret{}
	}

	if desiredKind.Spec.Keys.MaintenanceKeyPublic == nil {
		desiredKind.Spec.Keys.MaintenanceKeyPublic = &secret.Secret{}
	}

	return map[string]*secret.Secret{
		"bootstrapkeyprivate":   desiredKind.Spec.Keys.BootstrapKeyPrivate,
		"bootstrapkeypublic":    desiredKind.Spec.Keys.BootstrapKeyPublic,
		"maintenancekeyprivate": desiredKind.Spec.Keys.MaintenanceKeyPrivate,
		"maintenancekeypublic":  desiredKind.Spec.Keys.MaintenanceKeyPublic,
	}
}
