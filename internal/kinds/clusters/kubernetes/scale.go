package kubernetes

import (
	"fmt"
	"sync"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"

	"github.com/caos/orbiter/internal/core/helpers"
	"github.com/caos/orbiter/internal/core/operator/orbiter"
	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/kinds/clusters/kubernetes/edge/k8s"
	"github.com/caos/orbiter/logging"
)

type scaleablePool struct {
	pool         *pool
	desiredScale int
}

func ensureScale(
	logger logging.Logger,
	desired DesiredV0,
	currentComputes map[string]*Compute,
	nodeAgentsCurrent map[string]*orbiter.NodeAgentCurrent,
	nodeAgentsDesired map[string]*orbiter.NodeAgentSpec,
	kubeconfig *orbiter.Secret,
	controlplanePool *scaleablePool,
	workerPools []*scaleablePool,
	kubeAPI infra.Address,
	k8sVersion k8s.KubernetesVersion,
	k8sClient *k8s.Client) (done bool, err error) {

	newCurrentComputeCallback := func(tier Tier, poolSpec *poolSpec) func(infra.Compute) {
		return func(newCompute infra.Compute) {
			currentComputes[newCompute.ID()] = &Compute{
				Status: "maintaining",
				Metadata: ComputeMetadata{
					Tier:     tier,
					Provider: poolSpec.spec.Provider,
					Pool:     poolSpec.spec.Pool,
					Group:    poolSpec.group,
				},
			}
		}
	}

	done = true

	wCount := 0
	for _, w := range workerPools {
		wCount += w.desiredScale
	}
	logger.WithFields(map[string]interface{}{
		"control_plane_nodes": controlplanePool.desiredScale,
		"worker_nodes":        wCount,
	}).Debug("Ensuring scale")

	var wg sync.WaitGroup
	synchronizer := helpers.NewSynchronizer(&wg)
	wg.Add(1)
	go func() {
		delta := controlplanePool.desiredScale - len(controlplanePool.pool.computes())
		if delta >= 0 {
			synchronizer.Done(controlplanePool.pool.newComputes(
				delta,
				newCurrentComputeCallback(
					Controlplane,
					controlplanePool.pool.poolSpec)))
			return
		}

		//synchronizer.Done(errors.New("scaling down controlplane is not supported yet"))
		//return

		for _, compute := range controlplanePool.pool.computes()[controlplanePool.desiredScale:] {
			if goErr := k8sClient.DeleteNode(compute.ID(), compute, false); goErr != nil {
				synchronizer.Done(goErr)
				return
			}
		}
		synchronizer.Done(controlplanePool.pool.cleanupComputes())
	}()

	var mux sync.Mutex
	type downScaler struct {
		pool     pool
		computes []infra.Compute
	}
	scaleDownWorkers := make([]downScaler, 0)
	for _, wp := range workerPools {
		wg.Add(1)
		go func(workerPool *scaleablePool) {
			if k8sClient.Available() {
				if goErr := workerPool.pool.cleanupComputes(); err != nil {
					synchronizer.Done(goErr)
					return
				}
			}

			delta := workerPool.desiredScale - len(workerPool.pool.computes())
			if delta >= 0 {
				synchronizer.Done(workerPool.pool.newComputes(
					delta, newCurrentComputeCallback(
						Workers,
						workerPool.pool.poolSpec)))
				return
			}

			done = false
			mux.Lock()
			defer mux.Unlock()
			scaleDownWorkers = append(scaleDownWorkers, downScaler{
				pool:     *workerPool.pool,
				computes: workerPool.pool.computes()[workerPool.desiredScale:],
			})
			synchronizer.Done(nil)
		}(wp)
	}

	wg.Wait()

	if synchronizer.IsError() {
		return false, errors.Wrap(synchronizer, "failed to scale computes")
	}

	var joinCP infra.Compute
	var certsCP infra.Compute
	var joinWorkers infra.Computes

	computes := controlplanePool.pool.computes()
	var ensuredWorkers int
	ensuredControlplane := len(computes)
	for _, workerPool := range workerPools {
		workerComputes := workerPool.pool.computes()
		computes = append(computes, workerComputes...)
		ensuredWorkers += len(workerComputes)
	}

	cpIsReady := true
nodes:
	for _, compute := range computes {

		id := compute.ID()
		current := currentComputes[id]

		naDesired, ok := nodeAgentsDesired[compute.ID()]
		if !ok {
			naDesired = &orbiter.NodeAgentSpec{}
		}

		if naDesired.Software == nil {
			naDesired.Software = &orbiter.Software{}
		}

		if naDesired.Firewall == nil {
			naDesired.Firewall = &orbiter.Firewall{}
		}

		software := k8sVersion.DefineSoftware()
		naDesired.Software.Merge(software)

		fw := map[string]orbiter.Allowed{
			"kubelet": orbiter.Allowed{
				Port:     fmt.Sprintf("%d", 10250),
				Protocol: "tcp",
			},
		}

		if current.Metadata.Tier == Workers {
			fw["node-ports"] = orbiter.Allowed{
				Port:     fmt.Sprintf("%d-%d", 30000, 32767),
				Protocol: "tcp",
			}
		}

		if current.Metadata.Tier == Controlplane {
			fw["kubeapi-external"] = orbiter.Allowed{
				Port:     fmt.Sprintf("%d", kubeAPI.Port),
				Protocol: "tcp",
			}
			fw["kubeapi-internal"] = orbiter.Allowed{
				Port:     fmt.Sprintf("%d", 6666),
				Protocol: "tcp",
			}
			fw["etcd"] = orbiter.Allowed{
				Port:     fmt.Sprintf("%d-%d", 2379, 2380),
				Protocol: "tcp",
			}
			fw["kube-scheduler"] = orbiter.Allowed{
				Port:     fmt.Sprintf("%d", 10251),
				Protocol: "tcp",
			}
			fw["kube-controller"] = orbiter.Allowed{
				Port:     fmt.Sprintf("%d", 10252),
				Protocol: "tcp",
			}
		}

		if desired.Spec.Networking.Network == "calico" {
			fw["calico-bgp"] = orbiter.Allowed{
				Port:     fmt.Sprintf("%d", 179),
				Protocol: "tcp",
			}
		}

		firewall := orbiter.Firewall(fw)
		naDesired.Firewall.Merge(firewall)

		nodeIsJoining := false
		node, getNodeErr := k8sClient.GetNode(id)
		if getNodeErr == nil {
			nodeIsJoining = true
			for _, cond := range node.Status.Conditions {
				if cond.Type == v1.NodeReady {
					nodeIsJoining = false
					current.Status = "running"
					if current.Metadata.Tier == Controlplane {
						certsCP = compute
					}
					continue nodes
				}
			}
		}

		if current.Metadata.Tier == Controlplane && nodeIsJoining {
			cpIsReady = false
		}

		current.Status = "maintaining"
		done = false
		logger := logger.WithFields(map[string]interface{}{
			"compute": id,
			"tier":    current.Metadata.Tier,
		})

		if nodeIsJoining {
			logger.Info("Node is not ready yet")
		}

		naCurrent, ok := nodeAgentsCurrent[compute.ID()]
		if !ok {
			naCurrent = &orbiter.NodeAgentCurrent{} // Avoid many nil checks
		}

		nodeIsReady := naCurrent.NodeIsReady
		softwareIsReady := naCurrent.Software.Contains(software)
		firewallIsReady := naCurrent.Open.Contains(firewall)
		if !nodeIsReady || !softwareIsReady || !firewallIsReady {
			logger.WithFields(map[string]interface{}{
				"node":     nodeIsReady,
				"software": softwareIsReady,
				"firewall": firewallIsReady,
			}).Info("Compute is not ready to join yet")
			continue nodes
		}

		logger.Info("Compute is ready to join now")
		if current.Metadata.Tier == Controlplane && joinCP == nil {
			joinCP = compute
			continue nodes
		}
		joinWorkers = append(joinWorkers, compute)
	}

	if done {
		logger.WithFields(map[string]interface{}{
			"controlplane": ensuredControlplane,
			"workers":      ensuredWorkers,
		}).Debug("Scale is ensured")
		return true, nil
	}

	var jointoken string

	if certsCP != nil && (joinCP != nil || len(joinWorkers) > 0) {
		runes := []rune("abcdefghijklmnopqrstuvwxyz0123456789")
		jointoken = fmt.Sprintf("%s.%s", helpers.RandomStringRunes(6, runes), helpers.RandomStringRunes(16, runes))
		if _, err := certsCP.Execute(nil, nil, "sudo kubeadm token create "+jointoken); err != nil {
			return false, errors.Wrap(err, "creating new join token failed")
		}

		defer certsCP.Execute(nil, nil, "sudo kubeadm token delete "+jointoken)
	}

	var certKey []byte
	doKubeadmInit := certsCP == nil

	if joinCP != nil {

		if !doKubeadmInit && !cpIsReady {
			return false, nil
		}

		if !doKubeadmInit && certKey == nil {
			var err error
			certKey, err = certsCP.Execute(nil, nil, "sudo kubeadm init phase upload-certs --upload-certs | tail -1")
			if err != nil {
				return false, errors.Wrap(err, "uploading certs failed")
			}
			logger.Info("Refreshed certs")
		}

		joinKubeconfig, err := join(
			logger,
			joinCP,
			certsCP,
			desired,
			kubeAPI,
			jointoken,
			k8sVersion,
			string(certKey),
			true)

		if joinKubeconfig == nil || err != nil {
			return false, err
		}
		kubeconfig.Value = *joinKubeconfig
		return false, nil
	}

	if certsCP == nil {
		logger.Info("Awaiting controlplane initialization")
		return false, nil
	}

	for _, worker := range joinWorkers {
		wg.Add(1)
		go func(w infra.Compute) {
			_, goErr := join(
				logger,
				w,
				certsCP,
				desired,
				kubeAPI,
				string(jointoken),
				k8sVersion,
				"",
				false)
			synchronizer.Done(errors.Wrapf(goErr, "joining worker %s failed", w.ID()))
		}(worker)
	}

	wg.Wait()

	if synchronizer.IsError() {
		return false, errors.Wrap(synchronizer, "failed joining computes")
	}

	for _, down := range scaleDownWorkers {
		for _, cmp := range down.computes {
			if err := k8sClient.DeleteNode(cmp.ID(), cmp, true); err != nil {
				return false, errors.Wrapf(err, "failed deleting node %s from pool %s", cmp.ID(), down.pool.poolSpec.group)
			}
			//			defer func() {
			//				if err != nil {
			//					delete(curr.Computes, cmp.ID())
			//				}
			//			}()
		}
		if err := down.pool.cleanupComputes(); err != nil {
			return false, errors.Wrapf(err, "failed cleaning up computes %s from pool %s", infra.Computes(down.computes), down.pool.poolSpec.group)
		}
	}

	return false, nil
}
