package kubernetes

import (
	"fmt"

	"github.com/pkg/errors"
	v1 "k8s.io/api/core/v1"

	"github.com/caos/orbiter/internal/helpers"
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
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
	k8sVersion KubernetesVersion,
	k8sClient *Client,
	oneoff bool,
	initializeMachine func(infra.Machine, initializedPool) (initializedMachine, error)) (bool, error) {

	wCount := 0
	for _, w := range workerPools {
		wCount += w.desired.Nodes
	}
	logger.WithFields(map[string]interface{}{
		"control_plane_nodes": controlplanePool.desired.Nodes,
		"worker_nodes":        wCount,
	}).Debug("Ensuring scale")

	alignMachines := func(pool initializedPool) (initialized bool, err error) {

		existing, err := pool.machines()
		if err != nil {
			return false, err
		}
		delta := pool.desired.Nodes - len(existing)
		if delta > 0 {
			machines, err := newMachines(pool.infra, delta)
			if err != nil {
				return false, err
			}
			for _, machine := range machines {
				if _, err := initializeMachine(machine, pool); err != nil {
					return false, err
				}
			}
		} else {
			for _, machine := range existing[pool.desired.Nodes:] {
				if err := k8sClient.EnsureDeleted(machine.infra.ID(), machine.currentMachine, machine.infra, false); err != nil {
					return false, err
				}
				if err := machine.infra.Remove(); err != nil {
					return false, err
				}
			}
		}
		return delta <= 0, nil
	}

	upscalingDone, err := alignMachines(controlplanePool)
	if err != nil {
		return false, err
	}

	machines, err := controlplanePool.machines()
	if err != nil {
		return false, err
	}

	ensuredControlplane := len(machines)
	var ensuredWorkers int
	for _, workerPool := range workerPools {
		workerUpscalingDone, err := alignMachines(workerPool)
		if err != nil {
			return false, err
		}
		if !workerUpscalingDone {
			upscalingDone = false
		}

		workerMachines, err := workerPool.machines()
		if err != nil {
			return false, err
		}
		ensuredWorkers += len(workerMachines)
		machines = append(machines, workerMachines...)
	}

	if !upscalingDone {
		logger.Info(false, "Upscaled machines are not ready yet")
		return false, nil
	}

	var joinCP infra.Machine
	var certsCP infra.Machine
	var joinWorkers []initializedMachine
	cpIsReady := true
	done := true

nodes:
	for _, machine := range machines {

		id := machine.infra.ID()

		nodeIsJoining := false
		node, getNodeErr := k8sClient.GetNode(id)
		if getNodeErr == nil {
			nodeIsJoining = true
			for _, cond := range node.Status.Conditions {
				if cond.Type == v1.NodeReady {
					nodeIsJoining = false
					machine.currentMachine.Status.Kubernetes = "online"
					if machine.tier == Controlplane {
						certsCP = machine.infra
					}
					continue nodes
				}
			}
		}

		if machine.tier == Controlplane && nodeIsJoining {
			cpIsReady = false
		}

		done = false
		logger := logger.WithFields(map[string]interface{}{
			"machine": id,
			"tier":    machine.tier,
		})

		if nodeIsJoining {
			logger.Info(false, "Node is not ready yet")
		}

		if machine.tier == Controlplane && joinCP == nil {
			joinCP = machine.infra
			continue nodes
		}
		joinWorkers = append(joinWorkers, machine)
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

		if doKubeadmInit && (kubeconfig.Value != "" || !oneoff) {
			return false, errors.New("initializing a cluster is not supported when kubeconfig exists or the flag --recur is true")
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
			logger.Info(false, "Refreshed certs")
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
		return false, psf(logger.WithFields(map[string]interface{}{
			"type": "kubeconfig",
		}))
	}

	if certsCP == nil {
		logger.Info(false, "Awaiting controlplane initialization")
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
