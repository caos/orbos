package clusters

import (
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/orb"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func GetQueryAndDestroyFuncs(
	monitor mntr.Monitor,
	orb *orb.Orb,
	clusterID string,
	clusterTree *tree.Tree,
	orbiterCommit string,
	oneoff bool,
	deployOrbiter bool,
	clusterCurrent *tree.Tree,
	destroyProviders func() (map[string]interface{}, error),
	whitelistChan chan []*orbiter.CIDR,
	finishedChan chan struct{},
) (
	orbiter.QueryFunc,
	orbiter.DestroyFunc,
	bool,
	error,
) {

	switch clusterTree.Common.Kind {
	case "orbiter.caos.ch/KubernetesCluster":
		adaptFunc := func() (orbiter.QueryFunc, orbiter.DestroyFunc, bool, error) {
			return kubernetes.AdaptFunc(
				orb,
				clusterID,
				oneoff,
				deployOrbiter,
				destroyProviders,
				func(whitelist []*orbiter.CIDR) {
					go func() {
						monitor.Debug("Sending whitelist")
						whitelistChan <- whitelist
						close(whitelistChan)
					}()
					monitor.Debug("Whitelist sent")
				},
			)(
				monitor.WithFields(map[string]interface{}{"cluster": clusterID}),
				finishedChan,
				clusterTree,
				clusterCurrent,
			)
		}
		return orbiter.AdaptFuncGoroutine(adaptFunc)
		//				subassemblers[provIdx] = static.New(providerPath, generalOverwriteSpec, staticadapter.New(providermonitor, providerID, "/healthz", updatesDisabled, cfg.NodeAgent))
	default:
		return nil, nil, false, errors.Errorf("unknown cluster kind %s", clusterTree.Common.Kind)
	}
}

func GetSecrets(
	monitor mntr.Monitor,
	clusterTree *tree.Tree,
) (
	map[string]*secret.Secret,
	error,
) {

	switch clusterTree.Common.Kind {
	case "orbiter.caos.ch/KubernetesCluster":
		return kubernetes.SecretFunc()(
			monitor,
			clusterTree,
		)
	default:
		return nil, errors.Errorf("unknown cluster kind %s", clusterTree.Common.Kind)
	}
}

func RewriteMasterkey(
	monitor mntr.Monitor,
	newMasterkey string,
	clusterTree *tree.Tree,
) (
	map[string]*secret.Secret,
	error,
) {

	switch clusterTree.Common.Kind {
	case "orbiter.caos.ch/KubernetesCluster":
		return kubernetes.RewriteFunc(
			newMasterkey,
		)(
			monitor,
			clusterTree,
		)
	default:
		return nil, errors.Errorf("unknown cluster kind %s", clusterTree.Common.Kind)
	}
}
