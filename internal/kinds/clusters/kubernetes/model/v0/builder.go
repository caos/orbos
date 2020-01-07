package v0

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/kinds/clusters/kubernetes/model"
	"github.com/caos/orbiter/internal/kinds/providers/gce"
	gceadapter "github.com/caos/orbiter/internal/kinds/providers/gce/adapter"
	"github.com/caos/orbiter/internal/kinds/providers/static"
	staticadapter "github.com/caos/orbiter/internal/kinds/providers/static/adapter"
)

type subBuilder struct {
	assembler operator.Assembler
	desired   map[string]interface{}
	current   map[string]interface{}
}

func init() {
	build = func(desired map[string]interface{}, secrets *operator.Secrets, _ interface{}) (model.UserSpec, func(model.Config, []map[string]interface{}) (map[string]operator.Assembler, error)) {

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

				providerPath := []string{"deps", depID}
				generalOverwriteSpec := func(des map[string]interface{}) {
					if spec.Verbose {
						des["verbose"] = true
					}
				}

				kind, ok := depValue["kind"]
				if !ok {
					return nil, fmt.Errorf("Spec provider %+v has no kind field", depID)
				}
				kindStr, ok := kind.(string)
				if !ok {
					return nil, fmt.Errorf("Spec provider kind %v must be of type string", kind)
				}

				providerlogger := cfg.Params.Logger.WithFields(map[string]interface{}{
					"provider": depID,
				})
				providerID := cfg.Params.ID + depID
				switch kindStr {
				case "orbiter.caos.ch/GCEProvider":
					var lbs map[string]*infra.Ingress

					if !spec.Destroyed && spec.ControlPlane.Provider == depID {
						lbs = map[string]*infra.Ingress{
							"kubeapi": &infra.Ingress{
								Pools:            []string{spec.ControlPlane.Pool},
								HealthChecksPath: "/healthz",
							},
						}
					}
					subassemblers[depID] = gce.New(providerPath, generalOverwriteSpec, gceadapter.New(providerlogger, providerID, lbs, nil, "", cfg.Params.ConnectFromOutside))
				case "orbiter.caos.ch/StaticProvider":
					updatesDisabled := make([]string, 0)
					for _, pool := range spec.Workers {
						if pool.UpdatesDisabled {
							updatesDisabled = append(updatesDisabled, pool.Pool)
						}
					}

					if spec.ControlPlane.UpdatesDisabled {
						updatesDisabled = append(updatesDisabled, spec.ControlPlane.Pool)
					}

					subassemblers[depID] = static.New(providerPath, generalOverwriteSpec, staticadapter.New(providerlogger, providerID, "/healthz", updatesDisabled, cfg.NodeAgent))
				default:
					return nil, fmt.Errorf("Provider of kind %s is unknown", kindStr)
				}
			}

			return subassemblers, nil
		}
	}
}
