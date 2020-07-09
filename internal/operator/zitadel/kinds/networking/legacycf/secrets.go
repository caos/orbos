package legacycf

import (
	"github.com/caos/orbos/internal/operator/zitadel/kinds/networking/legacycf/config"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
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

		return getSecretsMap(desiredKind), nil
	}
}

func getSecretsMap(desiredKind *Desired) map[string]*secret.Secret {
	secrets := map[string]*secret.Secret{}
	if desiredKind.Spec == nil {
		desiredKind.Spec = &config.Config{}
	}

	if desiredKind.Spec.Credentials == nil {
		desiredKind.Spec.Credentials = &config.Credentials{}
	}

	if desiredKind.Spec.Credentials.User == nil {
		desiredKind.Spec.Credentials.User = &secret.Secret{}
	}

	if desiredKind.Spec.Credentials.APIKey == nil {
		desiredKind.Spec.Credentials.APIKey = &secret.Secret{}
	}

	secrets["credentials.user"] = desiredKind.Spec.Credentials.User
	secrets["credentials.apikey"] = desiredKind.Spec.Credentials.APIKey

	return secrets
}
