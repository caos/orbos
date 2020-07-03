package secretfuncs

import (
	"github.com/caos/orbos/internal/operator/boom/api"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
)

func GetSecrets() func(operator string) secret.Func {
	return func(operator string) secret.Func {
		if operator == "boom" {
			return api.SecretsFunc()
		} else if operator == "orbiter" {
			return orb.SecretsFunc()
		}

		return func(monitor mntr.Monitor, desiredTree *tree.Tree) (secrets map[string]*secret.Secret, err error) {
			return make(map[string]*secret.Secret), nil
		}
	}
}

func GetRewrite(masterkey string) func(operator string) secret.Func {
	return func(operator string) secret.Func {
		if operator == "boom" {
			return api.RewriteFunc(masterkey)
		} else if operator == "orbiter" {
			return orb.RewriteFunc(masterkey)
		}

		return func(monitor mntr.Monitor, desiredTree *tree.Tree) (secrets map[string]*secret.Secret, err error) {
			return make(map[string]*secret.Secret), nil
		}
	}
}
