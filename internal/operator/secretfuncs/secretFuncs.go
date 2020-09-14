package secretfuncs

import (
	"github.com/caos/orbos/internal/operator/boom/api"
	orborbiter "github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
	orbzitadel "github.com/caos/orbos/internal/operator/zitadel/kinds/orb"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
)

func GetSecrets() func(operator string) secret.Func {
	return func(operator string) secret.Func {
		if operator == "boom" {
			return api.SecretsFunc()
		} else if operator == "orbiter" {
			return orborbiter.SecretsFunc()
		} else if operator == "zitadel" {
			return orbzitadel.SecretsFunc()
		}

		return func(monitor mntr.Monitor, desiredTree *tree.Tree) (secrets map[string]*secret.Secret, err error) {
			return make(map[string]*secret.Secret), nil
		}
	}
}
