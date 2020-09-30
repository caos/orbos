package secretfuncs

import (
	"github.com/caos/orbos/internal/operator/boom/api"
	orbdb "github.com/caos/orbos/internal/operator/database/kinds/orb"
	orbnw "github.com/caos/orbos/internal/operator/networking/kinds/orb"
	orborbiter "github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
	"github.com/caos/orbos/mntr"
	secret2 "github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
)

func GetSecrets() func(operator string) secret2.Func {
	return func(operator string) secret2.Func {
		switch operator {
		case "boom":
			return api.SecretsFunc()
		case "orbiter":
			return orborbiter.SecretsFunc()
		case "database":
			return orbdb.SecretsFunc()
		case "networking":
			return orbnw.SecretsFunc()
		}

		return func(monitor mntr.Monitor, desiredTree *tree.Tree) (secrets map[string]*secret2.Secret, err error) {
			return make(map[string]*secret2.Secret), nil
		}
	}
}
