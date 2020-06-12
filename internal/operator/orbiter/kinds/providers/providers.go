package providers

import (
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/static"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func GetQueryAndDestroyFuncs(
	monitor mntr.Monitor,
	provID string,
	providerTree *tree.Tree,
	providerCurrent *tree.Tree,
	whitelistChan chan []*orbiter.CIDR,
	finishedChan chan bool,
) (
	orbiter.QueryFunc,
	orbiter.DestroyFunc,
	bool,
	error,
) {
	switch providerTree.Common.Kind {
	//			case "orbiter.caos.ch/GCEProvider":
	//				var lbs map[string]*infra.Ingress
	//
	//				if !kind.Spec.Destroyed && kind.Spec.ControlPlane.Provider == depID {
	//					lbs = map[string]*infra.Ingress{
	//						"kubeapi": &infra.Ingress{
	//							Pools:            []string{kind.Spec.ControlPlane.Pool},
	//							HealthChecksPath: "/healthz",
	//						},
	//					}
	//				}
	//				subassemblers[provIdx] = gce.New(providerPath, generalOverwriteSpec, gceadapter.New(providermonitor, providerID, lbs, nil, "", cfg.Params.ConnectFromOutside))
	case "orbiter.caos.ch/StaticProvider":
		//				updatesDisabled := make([]string, 0)
		//				for _, pool := range desiredKind.Spec.Workers {
		//					if pool.UpdatesDisabled {
		//						updatesDisabled = append(updatesDisabled, pool.Pool)
		//					}
		//				}
		//
		//				if desiredKind.Spec.ControlPlane.UpdatesDisabled {
		//					updatesDisabled = append(updatesDisabled, desiredKind.Spec.ControlPlane.Pool)
		//				}

		adaptFunc := func() (orbiter.QueryFunc, orbiter.DestroyFunc, bool, error) {
			return static.AdaptFunc(
				provID,
				func() []*orbiter.CIDR {
					monitor.Debug("Reading whitelist")
					return <-whitelistChan
				},
			)(
				monitor.WithFields(map[string]interface{}{"provider": provID}),
				finishedChan,
				providerTree,
				providerCurrent)
		}
		return orbiter.AdaptFuncGoroutine(adaptFunc)
	default:
		return nil, nil, false, errors.Errorf("unknown provider kind %s", providerTree.Common.Kind)
	}
}

func GetSecrets(
	monitor mntr.Monitor,
	providerTree *tree.Tree,
) (
	map[string]*secret.Secret,
	error,
) {
	switch providerTree.Common.Kind {
	case "orbiter.caos.ch/StaticProvider":
		return static.SecretsFunc()(
			monitor,
			providerTree,
		)
	default:
		return nil, errors.Errorf("unknown provider kind %s", providerTree.Common.Kind)
	}
}

func RewriteMasterkey(
	monitor mntr.Monitor,
	newMasterkey string,
	providerTree *tree.Tree,
) (
	map[string]*secret.Secret,
	error,
) {
	switch providerTree.Common.Kind {
	case "orbiter.caos.ch/StaticProvider":
		return static.RewriteFunc(
			newMasterkey,
		)(
			monitor,
			providerTree,
		)
	default:
		return nil, errors.Errorf("unknown provider kind %s", providerTree.Common.Kind)
	}
}
