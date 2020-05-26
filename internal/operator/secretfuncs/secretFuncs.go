package secretfuncs

import (
	"github.com/caos/orbos/internal/operator/boom/api"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
	orbconfig "github.com/caos/orbos/internal/orb"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
)

func Get(orbFile *orbconfig.Orb) func(operator string) secret.Func {
	return func(operator string) secret.Func {
		if operator == "boom" {
			return api.SecretsFunc(orbFile)
		} else if operator == "orbiter" {
			return orb.SecretsFunc(orbFile)
		}

		return func(monitor mntr.Monitor, desiredTree *tree.Tree) (secrets map[string]*secret.Secret, err error) {
			return make(map[string]*secret.Secret), nil
		}
	}
}
