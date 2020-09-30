package bucket

import (
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

		secrets = make(map[string]*secret2.Secret, 0)
		if desiredKind.Spec != nil {
			if desiredKind.Spec.ServiceAccountJSON == nil {
				desiredKind.Spec.ServiceAccountJSON = &secret2.Secret{}
			}
			secrets["serviceaccountjson"] = desiredKind.Spec.ServiceAccountJSON
		}
		return secrets, nil
	}
}
