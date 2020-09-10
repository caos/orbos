package operators

import (
	"errors"
	"github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/internal/git"
	boomapi "github.com/caos/orbos/internal/operator/boom/api"
	orbiterOrb "github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
	zitadelOrb "github.com/caos/orbos/internal/operator/zitadel/kinds/orb"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"strings"
)

const (
	boom    string = "boom"
	orbiter string = "orbiter"
	zitadel string = "zitadel"
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

			orbiterSecrets, err := orbiterOrb.SecretsFunc()(monitor, orbiterYML)
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

		foundZitadel, err := api.ExistsZitadelYml(gitClient)
		if err != nil {
			return nil, nil, err
		}
		if foundZitadel {
			zitadelYML, err := api.ReadZitadelYml(gitClient)
			if err != nil {
				return nil, nil, err
			}
			allTrees["zitadel"] = zitadelYML
			zitadelSecrets, err := zitadelOrb.SecretsFunc()(monitor, zitadelYML)
			if err != nil {
				return nil, nil, err
			}
			if zitadelSecrets != nil && len(zitadelSecrets) > 0 {
				for k, v := range zitadelSecrets {
					if k != "" && v != nil {
						allSecrets["zitadel."+k] = v
					}
				}
			}
		}

		return allSecrets, allTrees, nil
	}
}

func PushFunc() func(monitor mntr.Monitor, gitClient *git.Client, trees map[string]*tree.Tree, path string) error {
	return func(monitor mntr.Monitor, gitClient *git.Client, trees map[string]*tree.Tree, path string) error {
		operator := ""
		if strings.HasPrefix(path, orbiter) {
			operator = orbiter
		} else if strings.HasPrefix(path, boom) {
			operator = boom
		} else if strings.HasPrefix(path, zitadel) {
			operator = zitadel
		} else {
			return errors.New("Operator unknown")
		}

		desired, found := trees[operator]
		if !found {
			return errors.New("Operator file not found")
		}

		if operator == orbiter {
			return api.PushOrbiterDesiredFunc(gitClient, desired)(monitor)
		} else if operator == boom {
			return api.PushBoomDesiredFunc(gitClient, desired)(monitor)
		} else if operator == zitadel {
			return api.PushZitadelDesiredFunc(gitClient, desired)(monitor)
		}

		return errors.New("Operator push function unknown")
	}
}
