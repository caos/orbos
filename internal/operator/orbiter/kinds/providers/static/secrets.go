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

func RewriteFunc(newMasterkey string) secret.Func {
	return func(monitor mntr.Monitor, desiredTree *tree.Tree) (secrets map[string]*secret.Secret, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind, err := parseDesiredV0(desiredTree)
		if err != nil {
			return nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		secret.Masterkey = newMasterkey

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
	ret := make(map[string]*secret.Secret, 0)
	if desiredKind.Spec.Keys != nil {
		if desiredKind.Spec.Keys.BootstrapKeyPrivate != nil {
			ret["bootstrapkeyprivate"] = desiredKind.Spec.Keys.BootstrapKeyPrivate
		}
		if desiredKind.Spec.Keys.BootstrapKeyPrivate != nil {
			ret["bootstrapkeypublic"] = desiredKind.Spec.Keys.BootstrapKeyPublic
		}
		if desiredKind.Spec.Keys.BootstrapKeyPrivate != nil {
			ret["maintenancekeyprivate"] = desiredKind.Spec.Keys.MaintenanceKeyPrivate
		}
		if desiredKind.Spec.Keys.BootstrapKeyPrivate != nil {
			ret["maintenancekeypublic"] = desiredKind.Spec.Keys.MaintenanceKeyPublic
		}

	}
	return ret
}
