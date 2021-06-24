package kubernetes

import (
	"fmt"
	"strings"

	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	v1 "k8s.io/api/core/v1"
	macherrs "k8s.io/apimachinery/pkg/api/errors"
)

func scaleDown(pools []*initializedPool, k8sClient *kubernetes.Client, uninitializeMachine uninitializeMachineFunc, monitor mntr.Monitor, pdf func(mntr.Monitor) error) error {
	for _, pool := range pools {
		for _, machine := range pool.downscaling {
			id := machine.infra.ID()
			var existingK8sNode *v1.Node
			if k8sClient != nil {
				foundK8sNode, err := k8sClient.GetNode(id)
				if macherrs.IsNotFound(err) {
					err = nil
				}
				if err != nil {
					return fmt.Errorf("getting node %s from kube api failed: %w", id, err)
				}
				existingK8sNode = foundK8sNode
			}

			if existingK8sNode != nil {
				if err := k8sClient.Drain(machine.currentMachine, existingK8sNode, kubernetes.Deleting, true); err != nil {
					return err
				}
			}

			remove, err := machine.infra.Destroy()
			if err != nil {
				return err
			}

			monitor.Info("Resetting kubeadm")
			if _, resetErr := machine.infra.Execute(nil, "sudo kubeadm reset --force"); resetErr != nil {
				if !strings.Contains(resetErr.Error(), "command not found") {
					return resetErr
				}
			}

			if existingK8sNode != nil {
				if err := k8sClient.DeleteNode(id); err != nil {
				}
			}

			if !machine.currentMachine.GetUpdating() || machine.currentMachine.GetJoined() {
				machine.currentMachine.SetUpdating(true)
				machine.currentMachine.SetJoined(false)
				monitor.Changed("Node deleted")
			}

			uninitializeMachine(id)
			if req, _, unreq := machine.infra.ReplacementRequired(); req {
				unreq()
				pdf(monitor.WithFields(map[string]interface{}{
					"reason":   "unrequire machine replacement",
					"replaced": id,
				}))
			}
			if err := remove(); err != nil {
				return err
			}
			monitor.WithFields(map[string]interface{}{
				"machine": id,
				"tier":    machine.pool.tier,
			}).Changed("Machine removed")
		}
	}

	return nil
}
