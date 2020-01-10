package adapter

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/core/helpers"
	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/kinds/clusters/kubernetes/edge/k8s"
	"github.com/caos/orbiter/internal/kinds/clusters/kubernetes/model"
)

// TODO per pool:
// 1. Downscale if desired < current
// 2. Migrate
// 3. Upscale if desired > current
func ensureCluster(
	cfg *model.Config,
	curr *model.Current,
	providerPools map[string]map[string]infra.Pool,
	kubeAPIAddress infra.Address,
	secrets *operator.Secrets,
	k8sClient *k8s.Client) (err error) {

	kubeConfigKey := cfg.Params.ID + "_kubeconfig"
	joinTokenKey := cfg.Params.ID + "_jointoken"

	if !cfg.Spec.Destroyed && cfg.Spec.ControlPlane.Nodes != 1 && cfg.Spec.ControlPlane.Nodes != 3 && cfg.Spec.ControlPlane.Nodes != 5 {
		err = errors.New("Controlplane nodes can only be scaled to 1, 3 or 5")
		return err
	}

	var controlplanePool *scaleablePool
	var cpPoolComputes infra.Computes
	workerPools := make([]*scaleablePool, 0)
	workerComputes := make([]infra.Compute, 0)
	for providerName, provider := range providerPools {
		for poolName, wPool := range provider {
			if cfg.Spec.ControlPlane.Provider == providerName && cfg.Spec.ControlPlane.Pool == poolName {

				cpDesired := cfg.Spec.ControlPlane
				cpPool := providerPools[cpDesired.Provider][cpDesired.Pool]
				cfg.Params.Logger.WithFields(map[string]interface{}{
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
					newCurrentCompute(cfg, curr, comp, &model.ComputeMetadata{
						Tier:     model.Controlplane,
						Provider: cpDesired.Provider,
						Pool:     cpDesired.Pool,
						Group:    "",
					})
				}
				controlplanePool = &scaleablePool{
					pool: newPool(
						cfg,
						&poolSpec{group: "", spec: cpDesired},
						cpPool,
						k8sClient,
						cpPoolComputes),
					desiredScale: cpDesired.Nodes,
				}

				continue
			}
			var (
				wDesired *model.Pool
				group    string
			)
			for g, w := range cfg.Spec.Workers {
				if providerName == w.Provider && poolName == w.Pool {
					group = g
					wDesired = w
					break
				}
			}

			if wDesired == nil {
				wDesired = &model.Pool{
					Provider:        providerName,
					UpdatesDisabled: true,
					Nodes:           0,
					Pool:            poolName,
				}
			}

			cfg.Params.Logger.WithFields(map[string]interface{}{
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
					cfg,
					&poolSpec{group: group, spec: wDesired},
					wPool,
					k8sClient,
					wPoolComputes),
				desiredScale: wDesired.Nodes,
			})
			workerComputes = append(workerComputes, wPoolComputes...)
			for _, comp := range wPoolComputes {
				newCurrentCompute(cfg, curr, comp, &model.ComputeMetadata{
					Tier:     model.Workers,
					Provider: wDesired.Provider,
					Pool:     wDesired.Pool,
					Group:    group,
				})
			}
		}
	}

	if curr.Computes == nil {
		curr.Computes = make(map[string]*model.Compute)
	}

	if len(cpPoolComputes) == 0 {
		_ = secrets.Delete(kubeConfigKey)
		_ = secrets.Delete(joinTokenKey)
	} else {
		kc, _ := secrets.Read(kubeConfigKey)
		kcStrCast := string(kc)
		k8sClient.Refresh(&kcStrCast)
	}

	if cfg.Spec.Destroyed {
		var wg sync.WaitGroup
		synchronizer := helpers.NewSynchronizer(&wg)
		for _, compute := range append(cpPoolComputes, workerComputes...) {
			wg.Add(2)
			go func(cmp infra.Compute) {
				_, resetErr := cmp.Execute(nil, nil, "sudo kubeadm reset -f")
				_, rmErr := cmp.Execute(nil, nil, "sudo rm -rf /var/lib/etcd")
				synchronizer.Done(resetErr)
				synchronizer.Done(rmErr)
			}(compute)
		}
		wg.Wait()
		if synchronizer.IsError() {
			cfg.Params.Logger.Info(synchronizer.Error())
		}
		return nil
	}

	targetVersion := k8s.ParseString(cfg.Spec.Kubernetes)
	upgradingDone, err := ensureK8sVersion(
		cfg,
		targetVersion,
		k8sClient,
		curr.Computes,
		cpPoolComputes,
		workerComputes)
	if err != nil || !upgradingDone {
		cfg.Params.Logger.Debug("Upgrading is not done yet")
		return err
	}

	var scalingDone bool
	scalingDone, err = ensureScale(
		cfg,
		curr,
		secrets,
		kubeConfigKey,
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

func newCurrentCompute(cfg *model.Config, curr *model.Current, compute infra.Compute, meta *model.ComputeMetadata) {
	nodeagent := cfg.NodeAgent(compute)
	curr.Computes[compute.ID()] = &model.Compute{
		Status:    "maintaining",
		Metadata:  meta,
		Nodeagent: nodeagent,
	}
}
