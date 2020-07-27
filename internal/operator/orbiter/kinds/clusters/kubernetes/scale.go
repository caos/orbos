package kubernetes

import (
	"fmt"
	"sync"

	"github.com/caos/orbos/internal/secret"

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
	kubeAPI *infra.Address,
	k8sVersion KubernetesVersion,
	k8sClient *Client,
	oneoff bool,
	initializeMachine func(infra.Machine, initializedPool) initializedMachine,
	uninitializeMachine uninitializeMachineFunc) (changed bool, err error) {

	wCount := 0
	for _, w := range workerPools {
		wCount += w.desired.Nodes
	}
	monitor.WithFields(map[string]interface{}{
		"control_plane_nodes": controlplanePool.desired.Nodes,
		"worker_nodes":        wCount,
	}).Debug("Ensuring scale")

	var machines []*initializedMachine
	upscalingDone := true
	var wg sync.WaitGroup
	alignPool := func(pool initializedPool, ensured func(int)) {
		defer wg.Done()

		existing, alignErr := pool.machines()
		err = helpers.Concat(err, alignErr)
		delta := pool.desired.Nodes - len(existing)
		if delta > 0 {
			upscalingDone = false
			machines, alignErr := newMachines(pool.infra, delta)
			if alignErr != nil {
				err = helpers.Concat(err, alignErr)
				return
			}
			for _, machine := range machines {
				initializeMachine(machine, pool)
			}
		} else {
			for _, machine := range existing[pool.desired.Nodes:] {
				id := machine.infra.ID()
				delErr := k8sClient.EnsureDeleted(id, machine.currentMachine, machine.infra, false)
				if delErr != nil {
					err = helpers.Concat(err, delErr)
					return
				}

				rmErr := machine.infra.Remove()
				if rmErr != nil {
					err = helpers.Concat(err, rmErr)
					return
				}
				uninitializeMachine(id)
				monitor.WithFields(map[string]interface{}{
					"machine": id,
					"tier":    machine.pool.tier,
				}).Changed("Machine removed")
			}
		}

		if err != nil {
			return
		}
		poolMachines, listErr := pool.machines()
		if listErr != nil {
			err = helpers.Concat(err, listErr)
			return
		}
		machines = append(machines, poolMachines...)
		if ensured != nil {
			ensured(len(poolMachines))
		}
	}

	var ensuredControlplane int
	wg.Add(1)
	go alignPool(controlplanePool, func(ensured int) {
		ensuredControlplane = ensured
	})

	var ensuredWorkers int
	for _, workerPool := range workerPools {
		wg.Add(1)
		go alignPool(workerPool, func(ensured int) {
			ensuredWorkers += ensured
		})
	}
	wg.Wait()
	if err != nil {
		return false, err
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

		machineMonitor := monitor.WithFields(map[string]interface{}{
			"machine": machine.infra.ID(),
			"tier":    machine.pool.tier,
		})

		if machine.currentMachine.Unknown {
			machineMonitor.Info("Waiting for kubernetes node to leave unknown state before proceeding")
			return false, nil
		}

		isJoinedControlPlane := machine.pool.tier == Controlplane && machine.currentMachine.Joined

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

		if machine.pool.tier == Controlplane && joinCP == nil {
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
		if _, err := certsCP.Execute(nil, "sudo kubeadm token create "+jointoken); err != nil {
			return false, errors.Wrap(err, "creating new join token failed")
		}

		defer certsCP.Execute(nil, "sudo kubeadm token delete "+jointoken)

		if k8sVersion.equals(V1x18x0) {
			if _, err := certsCP.Execute(nil, "sudo kubeadm init phase bootstrap-token"); err != nil {
				return false, errors.Wrap(err, "Working around kubeadm bug failed, see https://kubernetes.io/docs/setup/production-environment/tools/kubeadm/troubleshooting-kubeadm/#not-possible-to-join-a-v1-18-node-to-a-v1-17-cluster-due-to-missing-rbac")
			}
		}
	}

	var certKey []byte
	doKubeadmInit := certsCP == nil

	if joinCP != nil {

		if doKubeadmInit && (desired.Spec.Kubeconfig != nil && desired.Spec.Kubeconfig.Value != "" || !oneoff) {
			return false, errors.New("initializing a cluster is not supported when kubeconfig exists or the flag --recur is passed")
		}

		if !doKubeadmInit && certKey == nil {
			var err error
			certKey, err = certsCP.Execute(nil, "sudo kubeadm init phase upload-certs --upload-certs | tail -1")
			if err != nil {
				return false, errors.Wrap(err, "uploading certs failed")
			}
			monitor.Info("Refreshed certs")
		}

		var joinKubeconfig *string
		joinKubeconfig, err = join(
			monitor,
			clusterID,
			joinCP,
			certsCP,
			*desired,
			kubeAPI,
			jointoken,
			k8sVersion,
			string(certKey))

		if err != nil {
			return false, err
		}

		if joinKubeconfig == nil || err != nil {
			return false, err
		}
		desired.Spec.Kubeconfig = &secret.Secret{Value: *joinKubeconfig}
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
			jointoken,
			k8sVersion,
			""); err != nil {
			return false, errors.Wrapf(err, "joining worker %s failed", worker.infra.ID())
		}
	}

	for _, pool := range append(workerPools, controlplanePool) {
		if err := pool.infra.EnsureMembers(); err != nil {
			return false, err
		}
	}

	return false, nil
}
