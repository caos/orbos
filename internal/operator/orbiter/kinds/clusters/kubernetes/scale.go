package kubernetes

import (
	"fmt"
	"github.com/caos/orbos/internal/api"
	"github.com/pkg/errors"

	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/mntr"
)

func ensureScale(
	monitor mntr.Monitor,
	clusterID string,
	desired *DesiredV0,
	psf api.SecretFunc,
	controlplanePool initializedPool,
	workerPools []initializedPool,
	kubeAPI infra.Address,
	k8sVersion KubernetesVersion,
	k8sClient *Client,
	oneoff bool,
	initializeMachine func(infra.Machine, initializedPool) (initializedMachine, error),
	uninitializeMachine uninitializeMachineFunc) (bool, error) {

	wCount := 0
	for _, w := range workerPools {
		wCount += w.desired.Nodes
	}
	monitor.WithFields(map[string]interface{}{
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
				id := machine.infra.ID()
				if err := k8sClient.EnsureDeleted(id, machine.currentMachine, machine.infra, false); err != nil {
					return false, err
				}
				if err := machine.infra.Remove(); err != nil {
					return false, err
				}
				uninitializeMachine(id)
				monitor.WithFields(map[string]interface{}{
					"machine": id,
					"tier":    machine.tier,
				}).Changed("Machine removed")
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
		monitor.Info("Upscaled machines are not ready yet")
		return false, nil
	}

	var joinCP *initializedMachine
	var certsCP infra.Machine
	var joinWorkers []*initializedMachine

nodes:
	for _, machine := range machines {

		isJoinedControlPlane := machine.tier == Controlplane && machine.currentMachine.Joined

		machineMonitor := monitor.WithFields(map[string]interface{}{
			"machine": machine.infra.ID(),
			"tier":    machine.tier,
		})

		if isJoinedControlPlane && machine.currentMachine.Online {
			certsCP = machine.infra
			continue nodes
		}

		if isJoinedControlPlane && !machine.currentMachine.Online {
			machineMonitor.Info("Awaiting controlplane to become ready")
			return false, nil
		}

		if machine.currentMachine.Online {
			continue nodes
		}

		if machine.currentMachine.Joined {
			machineMonitor.Info("Node is already joining")
			continue nodes
		}

		if machine.tier == Controlplane && joinCP == nil {
			joinCP = machine
			continue nodes
		}

		joinWorkers = append(joinWorkers, machine)
	}

	if joinCP == nil && len(joinWorkers) == 0 {
		monitor.WithFields(map[string]interface{}{
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

		if k8sVersion.equals(V1x18x0) {
			if _, err := certsCP.Execute(nil, nil, "sudo kubeadm init phase bootstrap-token"); err != nil {
				return false, errors.Wrap(err, "Working around kubeadm bug failed, see https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/troubleshooting-kubeadm/#not-possible-to-join-a-v1-18-node-to-a-v1-17-cluster-due-to-missing-rbac")
			}
		}
	}

	var certKey []byte
	doKubeadmInit := certsCP == nil

	if joinCP != nil {

		if doKubeadmInit && (desired.Spec.Kubeconfig.Value != "" || !oneoff) {
			return false, errors.New("initializing a cluster is not supported when kubeconfig exists or the flag --recur is passed")
		}

		if !doKubeadmInit && certKey == nil {
			var err error
			certKey, err = certsCP.Execute(nil, nil, "sudo kubeadm init phase upload-certs --upload-certs | tail -1")
			if err != nil {
				return false, errors.Wrap(err, "uploading certs failed")
			}
			monitor.Info("Refreshed certs")
		}

		joinKubeconfig, err := join(
			monitor,
			clusterID,
			joinCP,
			certsCP,
			*desired,
			kubeAPI,
			jointoken,
			k8sVersion,
			string(certKey))

		if joinKubeconfig == nil || err != nil {
			return false, err
		}
		desired.Spec.Kubeconfig.Value = *joinKubeconfig
		return false, psf(monitor.WithFields(map[string]interface{}{
			"type": "kubeconfig",
		}))
	}

	if certsCP == nil {
		monitor.Info("Awaiting controlplane initialization")
		return false, nil
	}

	for _, worker := range joinWorkers {
		if _, err := join(
			monitor,
			clusterID,
			worker,
			certsCP,
			*desired,
			kubeAPI,
			string(jointoken),
			k8sVersion,
			""); err != nil {
			return false, errors.Wrapf(err, "joining worker %s failed", worker.infra.ID())
		}
	}

	return false, nil
}
