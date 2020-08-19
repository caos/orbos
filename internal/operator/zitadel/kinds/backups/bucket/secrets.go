package bucket

import (
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

		secrets = make(map[string]*secret.Secret, 0)
		if desiredKind.Spec != nil {
			if desiredKind.Spec.ServiceAccountJSON == nil {
				desiredKind.Spec.ServiceAccountJSON = &secret.Secret{}
			}
			secrets["serviceaccountjson"] = desiredKind.Spec.ServiceAccountJSON
		}
		return secrets, nil
	}
}
