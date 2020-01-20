package kubernetes

import (
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/kubernetes/edge/k8s"
	"github.com/caos/orbiter/logging"
)

// TODO per pool:
// 1. Downscale if desired < current
// 2. Migrate
// 3. Upscale if desired > current
func ensureCluster(
	logger logging.Logger,
	desired DesiredV0,
	curr *CurrentCluster,
	nodeAgentsCurrent map[string]*common.NodeAgentCurrent,
	nodeAgentsDesired map[string]*common.NodeAgentSpec,
	providerPools map[string]map[string]infra.Pool,
	kubeAPIAddress infra.Address,
	kubeconfig *orbiter.Secret,
	psf orbiter.PushSecretsFunc,
	k8sClient *k8s.Client,
	repoURL string,
	repoKey string,
	orbiterCommit string) (err error) {

	if desired.Spec.ControlPlane.Nodes != 1 && desired.Spec.ControlPlane.Nodes != 3 && desired.Spec.ControlPlane.Nodes != 5 {
		err = errors.New("Controlplane nodes can only be scaled to 1, 3 or 5")
		return err
	}

	var controlplanePool *scaleablePool
	var cpPoolComputes infra.Computes
	workerPools := make([]*scaleablePool, 0)
	workerComputes := make([]infra.Compute, 0)
	for providerName, provider := range providerPools {
		for poolName, wPool := range provider {
			if desired.Spec.ControlPlane.Provider == providerName && desired.Spec.ControlPlane.Pool == poolName {

				cpDesired := desired.Spec.ControlPlane
				cpPool := providerPools[cpDesired.Provider][cpDesired.Pool]
				logger.WithFields(map[string]interface{}{
					"provider": cpDesired.Provider,
					"pool":     cpDesired.Pool,
					"tier":     "controlplane",
					"address":  cpPool,
				}).Debug("Using for pool")
				cpPoolComputes, err = cpPool.GetComputes(true)
				if err != nil {
					return err
				}
				for _, comp := range cpPoolComputes {
					curr.Computes[comp.ID()] = &Compute{
						Status: "maintaining",
						Metadata: ComputeMetadata{
							Tier:     Controlplane,
							Provider: cpDesired.Provider,
							Pool:     cpDesired.Pool,
						},
					}
					nodeAgentsDesired[comp.ID()] = &common.NodeAgentSpec{
						ChangesAllowed: !cpDesired.UpdatesDisabled,
						Software:       &common.Software{},
						Firewall:       &common.Firewall{},
					}
				}
				controlplanePool = &scaleablePool{
					pool: newPool(
						logger,
						repoURL,
						repoKey,
						&poolSpec{group: "", spec: cpDesired},
						cpPool,
						k8sClient,
						cpPoolComputes),
					desiredScale: cpDesired.Nodes,
				}

				continue
			}
			var (
				wDesired *Pool
				group    string
			)
			for g, w := range desired.Spec.Workers {
				if providerName == w.Provider && poolName == w.Pool {
					group = g
					wDesired = w
					break
				}
			}

			if wDesired == nil {
				wDesired = &Pool{
					Provider:        providerName,
					UpdatesDisabled: true,
					Nodes:           0,
					Pool:            poolName,
				}
			}

			logger.WithFields(map[string]interface{}{
				"provider": wDesired.Provider,
				"pool":     wDesired.Pool,
				"tier":     "workers",
				"address":  wPool,
			}).Debug("Searching for pool")
			var wPoolComputes []infra.Compute
			wPoolComputes, err = wPool.GetComputes(true)
			if err != nil {
				return err
			}
			workerPools = append(workerPools, &scaleablePool{
				pool: newPool(
					logger,
					repoURL,
					repoKey,
					&poolSpec{group: group, spec: *wDesired},
					wPool,
					k8sClient,
					wPoolComputes),
				desiredScale: wDesired.Nodes,
			})
			workerComputes = append(workerComputes, wPoolComputes...)
			for _, comp := range wPoolComputes {
				curr.Computes[comp.ID()] = &Compute{
					Status: "maintaining",
					Metadata: ComputeMetadata{
						Tier:     Workers,
						Provider: wDesired.Provider,
						Pool:     wDesired.Pool,
						Group:    group,
					},
				}
				nodeAgentsDesired[comp.ID()] = &common.NodeAgentSpec{
					ChangesAllowed: !wDesired.UpdatesDisabled,
					Software:       &common.Software{},
					Firewall:       &common.Firewall{},
				}
			}
		}
	}

	if curr.Computes == nil {
		curr.Computes = make(map[string]*Compute)
	}

	if kubeconfig != nil && kubeconfig.Value != "" {
		k8sClient.Refresh(&kubeconfig.Value)
	}

	targetVersion := k8s.ParseString(desired.Spec.Versions.Kubernetes)
	upgradingDone, err := ensureK8sVersion(
		logger,
		orbiterCommit,
		repoURL,
		repoKey,
		targetVersion,
		k8sClient,
		curr.Computes,
		nodeAgentsCurrent,
		nodeAgentsDesired,
		cpPoolComputes,
		workerComputes)
	if err != nil || !upgradingDone {
		logger.Debug("Upgrading is not done yet")
		return err
	}

	var scalingDone bool
	scalingDone, err = ensureScale(
		logger,
		desired,
		curr.Computes,
		nodeAgentsCurrent,
		nodeAgentsDesired,
		kubeconfig,
		psf,
		controlplanePool,
		workerPools,
		kubeAPIAddress,
		targetVersion,
		k8sClient)
	if err != nil {
		return err
	}

	if scalingDone {
		curr.Status = "running"
	}

	return nil
}
