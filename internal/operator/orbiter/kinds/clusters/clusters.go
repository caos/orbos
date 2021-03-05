package clusters

import (
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/git"
	"github.com/caos/orbos/pkg/labels"
	"github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
)

func GetQueryAndDestroyFuncs(
	monitor mntr.Monitor,
	operator *labels.Operator,
	clusterID string,
	clusterTree *tree.Tree,
	oneoff bool,
	pprof bool,
	deployOrbiter bool,
	clusterCurrent *tree.Tree,
	destroyProviders func(map[string]interface{}) (map[string]interface{}, error),
	whitelistChan chan []*orbiter.CIDR,
	finishedChan chan struct{},
	gitClient *git.Client,
) (
	orbiter.QueryFunc,
	orbiter.DestroyFunc,
	orbiter.ConfigureFunc,
	bool,
	map[string]*secret.Secret,
	error,
) {

	switch clusterTree.Common.Kind {
	case "orbiter.caos.ch/KubernetesCluster":
		return kubernetes.AdaptFunc(
			labels.MustForAPI(operator, "KubernetesCluster", clusterTree.Common.Version),
			clusterID,
			oneoff,
			deployOrbiter,
			pprof,
			destroyProviders,
			func(whitelist []*orbiter.CIDR) {
				go func() {
					monitor.Debug("Sending whitelist")
					whitelistChan <- whitelist
					close(whitelistChan)
				}()
				monitor.Debug("Whitelist sent")
			},
			gitClient,
		)(
			monitor.WithFields(map[string]interface{}{"cluster": clusterID}),
			finishedChan,
			clusterTree,
			clusterCurrent,
		)

		//				subassemblers[provIdx] = static.New(providerPath, generalOverwriteSpec, staticadapter.New(providermonitor, providerID, "/healthz", updatesDisabled, cfg.NodeAgent))
	default:
		return nil, nil, nil, false, nil, errors.Errorf("unknown cluster kind %s", clusterTree.Common.Kind)
	}
}
