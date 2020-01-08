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
	build = func(serialized map[string]interface{}, secrets *operator.Secrets, _ interface{}) (model.UserSpec, func(model.Config) ([]operator.Assembler, error)) {

		kind := struct {
			Spec model.UserSpec
			Deps struct {
				Providers []map[string]interface{}
			}
		}{}
		err := mapstructure.Decode(serialized, &kind)

		return kind.Spec, func(cfg model.Config) ([]operator.Assembler, error) {

			if err != nil {
				return nil, err
			}

			subassemblers := make([]operator.Assembler, len(kind.Deps.Providers))
			for provIdx, depValue := range kind.Deps.Providers {

				depIDIface, ok := depValue["id"]
				if !ok {
					return nil, errors.Errorf("dependency %+v has no id", depValue)
				}

				depID := fmt.Sprintf("%v", depIDIface)

				providerPath := []string{"deps", "providers", depID}
				generalOverwriteSpec := func(des map[string]interface{}) {
					if kind.Spec.Verbose {
						des["verbose"] = true
					}
				}

				provKind, ok := depValue["kind"]
				if !ok {
					return nil, fmt.Errorf("Spec provider %+v has no kind field", depID)
				}
				kindStr, ok := provKind.(string)
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

					if !kind.Spec.Destroyed && kind.Spec.ControlPlane.Provider == depID {
						lbs = map[string]*infra.Ingress{
							"kubeapi": &infra.Ingress{
								Pools:            []string{kind.Spec.ControlPlane.Pool},
								HealthChecksPath: "/healthz",
							},
						}
					}
					subassemblers[provIdx] = gce.New(providerPath, generalOverwriteSpec, gceadapter.New(providerlogger, providerID, lbs, nil, "", cfg.Params.ConnectFromOutside))
				case "orbiter.caos.ch/StaticProvider":
					updatesDisabled := make([]string, 0)
					for _, pool := range kind.Spec.Workers {
						if pool.UpdatesDisabled {
							updatesDisabled = append(updatesDisabled, pool.Pool)
						}
					}

					if kind.Spec.ControlPlane.UpdatesDisabled {
						updatesDisabled = append(updatesDisabled, kind.Spec.ControlPlane.Pool)
					}

					subassemblers[provIdx] = static.New(providerPath, generalOverwriteSpec, staticadapter.New(providerlogger, providerID, "/healthz", updatesDisabled, cfg.NodeAgent))
				default:
					return nil, fmt.Errorf("Provider of kind %s is unknown", kindStr)
				}
			}

			return subassemblers, nil
		}
	}
}
