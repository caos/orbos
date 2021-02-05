package kubernetes

import (
	"github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
)

func scaleDown(pools []*initializedPool, k8sClient *kubernetes.Client, uninitializeMachine uninitializeMachineFunc, monitor mntr.Monitor, pdf api.PushDesiredFunc) error {
	for _, pool := range pools {
		for _, machine := range pool.downscaling {
			id := machine.infra.ID()
			if err := k8sClient.EnsureDeleted(id, machine.currentMachine, machine.infra); err != nil {
				return err
			}
			uninitializeMachine(id)
			if req, _, unreq := machine.infra.ReplacementRequired(); req {
				unreq()
				pdf(monitor.WithFields(map[string]interface{}{
					"reason":   "unrequire machine replacement",
					"replaced": id,
				}))
			}
			if err := machine.infra.Remove(); err != nil {
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
