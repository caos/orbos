package v0

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/kinds/loadbalancers/dynamic"
	dynamiclbadapter "github.com/caos/orbiter/internal/kinds/loadbalancers/dynamic/adapter"
	"github.com/caos/orbiter/internal/kinds/loadbalancers/external"
	externallbadapter "github.com/caos/orbiter/internal/kinds/loadbalancers/external/adapter"
	"github.com/caos/orbiter/internal/kinds/providers/static/model"
)

func init() {
	build = func(desired map[string]interface{}, _ *operator.Secrets, _ interface{}) (model.UserSpec, func(model.Config, []map[string]interface{}) (map[string]operator.Assembler, error)) {

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

				generalOverwriteSpec := func(des map[string]interface{}) {
					if spec.Verbose {
						des["verbose"] = true
					}
				}
				/*
					lbLogger := cfg.Logger.WithFields(map[string]interface{}{
						"lb": depKey,
					})
				*/
				depPath := []string{"deps", depID}
				depKind := depValue["kind"]
				/*				ingress, ok := cfg.Ingresses[depKey]
								if !ok {
									continue
								}
				*/
				switch depKind {
				case "orbiter.caos.ch/ExternalLoadBalancer":
					subassemblers[depID] = external.New(depPath, generalOverwriteSpec, externallbadapter.New())
				case "orbiter.caos.ch/DynamicLoadBalancer":
					subassemblers[depID] = dynamic.New(depPath, generalOverwriteSpec, dynamiclbadapter.New(spec.RemoteUser))
				default:
					return subassemblers, errors.Errorf("unknown dependency type %s", depKind)
				}
			}

			return subassemblers, nil
		}
	}
}
