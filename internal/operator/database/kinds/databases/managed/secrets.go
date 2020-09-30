package managed

import (
	"github.com/caos/orbos/internal/operator/database/kinds/backups"
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
		desiredKind := &DesiredV0{Common: desiredTree.Common}
		if err := desiredTree.Original.Decode(desiredKind); err != nil {
			return nil, errors.Wrapf(err, "unmarshaling desired state for kind %s failed", desiredTree.Common.Kind)
		}

		desiredTree.Parsed = desiredKind

		allSecrets := make(map[string]*secret2.Secret)
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

func appendSecrets(prefix string, into, add map[string]*secret2.Secret) {
	for key, secret := range add {
		name := prefix + "." + key
		into[name] = secret
	}
}
