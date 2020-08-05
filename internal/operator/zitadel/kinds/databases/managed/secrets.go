package managed

import (
	"github.com/caos/orbos/internal/operator/zitadel/kinds/backups"
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
		desiredKind := &DesiredV0{Common: desiredTree.Common}
		if err := desiredTree.Original.Decode(desiredKind); err != nil {
			return nil, errors.Wrapf(err, "unmarshaling desired state for kind %s failed", desiredTree.Common.Kind)
		}

		desiredTree.Parsed = desiredKind

		allSecrets := make(map[string]*secret.Secret)
		for k, v := range desiredKind.Spec.Backups {
			backupSecrets, err := backups.GetSecrets(monitor, v)
			if err != nil {
				return nil, err
			}

			appendSecrets(k, allSecrets, backupSecrets)
		}
		return allSecrets, nil
	}
}

func appendSecrets(prefix string, into, add map[string]*secret.Secret) {
	for key, secret := range add {
		name := prefix + "." + key
		into[name] = secret
	}
}
