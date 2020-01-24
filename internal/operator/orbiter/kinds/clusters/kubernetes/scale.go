package kubernetes

import (
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"

	"github.com/caos/orbiter/internal/helpers"
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/kubernetes/edge/k8s"
	"github.com/caos/orbiter/logging"
)

func ensureScale(
	logger logging.Logger,
	desired DesiredV0,
	kubeconfig *orbiter.Secret,
	psf orbiter.PushSecretsFunc,
	controlplanePool initializedPool,
	workerPools []initializedPool,
	kubeAPI infra.Address,
	k8sVersion k8s.KubernetesVersion,
	k8sClient *k8s.Client,
	oneoff bool,
	initializeCompute func(infra.Compute, initializedPool) (initializedCompute, error)) (bool, error) {

	wCount := 0
	for _, w := range workerPools {
		wCount += w.desired.Nodes
	}
	logger.WithFields(map[string]interface{}{
		"control_plane_nodes": controlplanePool.desired.Nodes,
		"worker_nodes":        wCount,
	}).Debug("Ensuring scale")

	alignComputes := func(pool initializedPool) (err error) {

		existing, err := pool.computes()
		if err != nil {
			return err
		}
		delta := pool.desired.Nodes - len(existing)
		if delta > 0 {
			computes, err := newComputes(pool.infra, delta)
			if err != nil {
				return err
			}
			for _, compute := range computes {
				if _, err := initializeCompute(compute, pool); err != nil {
					return err
				}
			}
		} else {
			for _, compute := range existing[pool.desired.Nodes:] {
				if err := k8sClient.EnsureDeleted(compute.infra.ID(), compute.infra, false); err != nil {
					return err
				}
				if err := compute.infra.Remove(); err != nil {
					return err
				}
			}
		}
		return nil
	}

	if err := alignComputes(controlplanePool); err != nil {
		return false, err
	}

	computes, err := controlplanePool.computes()
	if err != nil {
		return false, err
	}

	ensuredControlplane := len(computes)
	var ensuredWorkers int
	for _, workerPool := range workerPools {
		if err := alignComputes(workerPool); err != nil {
			return false, err
		}

		workerComputes, err := workerPool.computes()
		if err != nil {
			return false, err
		}
		ensuredWorkers += len(workerComputes)
		computes = append(computes, workerComputes...)
	}

	var joinCP infra.Compute
	var certsCP infra.Compute
	var joinWorkers []initializedCompute
	cpIsReady := true
	done := true

nodes:
	for _, compute := range computes {

		id := compute.infra.ID()

		nodeIsJoining := false
		node, getNodeErr := k8sClient.GetNode(id)
		if getNodeErr == nil {
			nodeIsJoining = true
			for _, cond := range node.Status.Conditions {
				if cond.Type == v1.NodeReady {
					nodeIsJoining = false
					compute.markAsRunning()
					if compute.tier == Controlplane {
						certsCP = compute.infra
					}
					continue nodes
				}
			}
		}

		if compute.tier == Controlplane && nodeIsJoining {
			cpIsReady = false
		}

		done = false
		logger := logger.WithFields(map[string]interface{}{
			"compute": id,
			"tier":    compute.tier,
		})

		if nodeIsJoining {
			logger.Info("Node is not ready yet")
		}

		if compute.tier == Controlplane && joinCP == nil {
			joinCP = compute.infra
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

		if doKubeadmInit && !oneoff {
			return false, errors.New("initializing a cluster is not supported when flag --recur is passed")
		}

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
		return false, psf()
	}

	if certsCP == nil {
		logger.Info("Awaiting controlplane initialization")
		return false, nil
	}

	for _, worker := range joinWorkers {
		if _, err := join(
			logger,
			worker.infra,
			certsCP,
			desired,
			kubeAPI,
			string(jointoken),
			k8sVersion,
			"",
			false); err != nil {
			return false, errors.Wrapf(err, "joining worker %s failed", worker.infra.ID())
		}
	}

	return false, nil
}
