package orb

import (
	"github.com/caos/orbos/internal/operator/networking/kinds/networking"
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

		desiredKind, err := ParseDesiredV0(desiredTree)
		if err != nil {
			return nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		return networking.GetSecrets(monitor, desiredKind.Networking)
	}
}
