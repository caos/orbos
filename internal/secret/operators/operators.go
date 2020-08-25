package operators

import (
	"github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/internal/git"
	boomapi "github.com/caos/orbos/internal/operator/boom/api"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
)

func GetAllSecretsFunc() func(monitor mntr.Monitor, gitClient *git.Client) (map[string]*secret.Secret, map[string]*tree.Tree, error) {
	return func(monitor mntr.Monitor, gitClient *git.Client) (map[string]*secret.Secret, map[string]*tree.Tree, error) {

		monitor.Info("Reading all secrets")
		allSecrets := make(map[string]*secret.Secret, 0)
		allTrees := make(map[string]*tree.Tree, 0)
		foundBoom, err := api.ExistsBoomYml(gitClient)
		if err != nil {
			return nil, nil, err
		}

		if foundBoom {
			boomYML, err := api.ReadBoomYml(gitClient)
			if err != nil {
				return nil, nil, err
			}
			allTrees["boom"] = boomYML

			boomSecrets, err := boomapi.SecretsFunc()(monitor, boomYML)
			if err != nil {
				return nil, nil, err
			}

			if boomSecrets != nil && len(boomSecrets) > 0 {
				for k, v := range boomSecrets {
					if k != "" && v != nil {
						allSecrets["boom."+k] = v
					}
				}
			}
		}

		foundOrbiter, err := api.ExistsOrbiterYml(gitClient)
		if err != nil {
			return nil, nil, err
		}
		if foundOrbiter {
			orbiterYML, err := api.ReadOrbiterYml(gitClient)
			if err != nil {
				return nil, nil, err
			}
			allTrees["orbiter"] = orbiterYML

			orbiterSecrets, err := orb.SecretsFunc()(monitor, orbiterYML)
			if err != nil {
				return nil, nil, err
			}

			if orbiterSecrets != nil && len(orbiterSecrets) > 0 {
				for k, v := range orbiterSecrets {
					if k != "" && v != nil {
						allSecrets["orbiter."+k] = v
					}
				}
			}
		}

		return allSecrets, allTrees, nil
	}
}
