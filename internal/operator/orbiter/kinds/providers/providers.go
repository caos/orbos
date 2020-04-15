package providers

import (
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/static"
	"github.com/caos/orbiter/internal/orb"
	"github.com/caos/orbiter/internal/secret"
	"github.com/caos/orbiter/internal/tree"
	"github.com/caos/orbiter/mntr"
	"github.com/pkg/errors"
)

func GetQueryAndDestroyFuncs(
	monitor mntr.Monitor,
	orb *orb.Orb,
	provID string,
	providerTree *tree.Tree,
	providerCurrent *tree.Tree,
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

		return static.AdaptFunc(
			orb.Masterkey,
			provID,
		)(
			monitor.WithFields(map[string]interface{}{"provider": provID}),
			providerTree,
			providerCurrent)
	default:
		return nil, nil, false, errors.Errorf("unknown provider kind %s", providerTree.Common.Kind)
	}

}

func GetSecrets(
	monitor mntr.Monitor,
	masterkey string,
	providerTree *tree.Tree,
) (
	map[string]*secret.Secret,
	error,
) {
	switch providerTree.Common.Kind {
	case "orbiter.caos.ch/StaticProvider":
		return static.SecretsFunc(
			masterkey,
		)(
			monitor,
			providerTree,
		)
	default:
		return nil, errors.Errorf("unknown provider kind %s", providerTree.Common.Kind)
	}
}
