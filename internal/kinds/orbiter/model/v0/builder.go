package v0

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/kinds/clusters/kubernetes"
	"github.com/caos/orbiter/internal/kinds/clusters/kubernetes/adapter"
	k8s "github.com/caos/orbiter/internal/kinds/clusters/kubernetes/model"
	"github.com/caos/orbiter/internal/kinds/orbiter/model"
)

func init() {
	build = func(desired map[string]interface{}, _ *operator.Secrets, dependant interface{}) (model.UserSpec, func(model.Config, []map[string]interface{}) (map[string]operator.Assembler, error)) {

		spec := model.UserSpec{}
		err := mapstructure.Decode(desired, &spec)

		return spec, func(cfg model.Config, deps []map[string]interface{}) (map[string]operator.Assembler, error) {

			if err != nil {
				return nil, err
			}
			subassemblers := make(map[string]operator.Assembler)
			for _, depValue := range deps {

				depIDIface, ok := depValue["id"]
				if !ok {
					return nil, errors.Errorf("dependency %+v has no id", depValue)
				}

				depID := fmt.Sprintf("%v", depIDIface)

				overwriteDesired := func(des map[string]interface{}) {
					if spec.Verbose {
						des["verbose"] = true
					}
					if spec.Destroyed {
						des["destroyed"] = true
					}
				}
				kind, ok := depValue["kind"]
				if !ok {
					return nil, fmt.Errorf("Desired cluster %+v has no kind field", depID)
				}

				kindStr, ok := kind.(string)
				if !ok {
					return nil, fmt.Errorf("Desired cluster kind %v must be of type string", kind)
				}

				path := []string{"deps", depID}
				switch kindStr {
				case "orbiter.caos.ch/KubernetesCluster":
					subassemblers[depID] = kubernetes.New(path, overwriteDesired, adapter.New(k8s.Parameters{
						Logger: cfg.Logger.WithFields(map[string]interface{}{
							"cluster": depID,
						}),
						ID:                 cfg.ConfigID + depID,
						SelfAbsolutePath:   path,
						RepoURL:            cfg.NodeagentRepoURL,
						RepoKey:            cfg.NodeagentRepoKey,
						MasterKey:          cfg.Masterkey,
						OrbiterVersion:     cfg.OrbiterVersion,
						OrbiterCommit:      cfg.OrbiterCommit,
						CurrentFile:        cfg.CurrentFile,
						SecretsFile:        cfg.SecretsFile,
						ConnectFromOutside: cfg.ConnectFromOutside,
					}))
				default:
					return nil, fmt.Errorf("Unknown cluster kind %s", kindStr)
				}
			}
			return subassemblers, nil
		}
	}
}
