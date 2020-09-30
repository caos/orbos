package dynamic

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
		desiredKind := &Desired{Common: desiredTree.Common}
		if err := desiredTree.Original.Decode(desiredKind); err != nil {
			return nil, errors.Wrapf(err, "unmarshaling desired state for kind %s failed", desiredTree.Common.Kind)
		}

		desiredTree.Parsed = desiredKind

		return make(map[string]*secret2.Secret, 0), nil
	}
}
