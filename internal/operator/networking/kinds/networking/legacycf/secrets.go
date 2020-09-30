package legacycf

import (
	"github.com/caos/orbos/internal/operator/networking/kinds/networking/legacycf/config"
	"github.com/caos/orbos/mntr"
	secret2 "github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
)

func SecretsFunc() secret2.Func {
	return func(monitor mntr.Monitor, desiredTree *tree.Tree) (secrets map[string]*secret2.Secret, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		desiredKind, err := parseDesired(desiredTree)
		if err != nil {
			return nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		return getSecretsMap(desiredKind), nil
	}
}

func getSecretsMap(desiredKind *Desired) map[string]*secret2.Secret {
	secrets := map[string]*secret2.Secret{}
	if desiredKind.Spec == nil {
		desiredKind.Spec = &config.ExternalConfig{}
	}

	if desiredKind.Spec.Credentials == nil {
		desiredKind.Spec.Credentials = &config.Credentials{}
	}

	if desiredKind.Spec.Credentials.User == nil {
		desiredKind.Spec.Credentials.User = &secret2.Secret{}
	}

	if desiredKind.Spec.Credentials.APIKey == nil {
		desiredKind.Spec.Credentials.APIKey = &secret2.Secret{}
	}

	if desiredKind.Spec.Credentials.UserServiceKey == nil {
		desiredKind.Spec.Credentials.UserServiceKey = &secret2.Secret{}
	}

	secrets["credentials.user"] = desiredKind.Spec.Credentials.User
	secrets["credentials.apikey"] = desiredKind.Spec.Credentials.APIKey
	secrets["credentials.userservicekey"] = desiredKind.Spec.Credentials.UserServiceKey

	return secrets
}
