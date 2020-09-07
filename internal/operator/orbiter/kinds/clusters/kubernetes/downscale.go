package kubernetes

import (
	"github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/mntr"
)

func scaleDown(pools []*initializedPool, k8sClient *Client, uninitializeMachine uninitializeMachineFunc, monitor mntr.Monitor, pdf api.PushDesiredFunc) error {
	var unrequired []string
	for _, pool := range pools {
		for _, machine := range pool.downscaling {
			id := machine.infra.ID()
			if err := k8sClient.EnsureDeleted(id, machine.currentMachine, machine.infra); err != nil {
				return err
			}

			if err := machine.infra.Remove(); err != nil {
				return err
			}
			if req, _, unreq := machine.infra.ReplacementRequired(); req {
				unreq()
				unrequired = append(unrequired, id)
			}
			uninitializeMachine(id)
			monitor.WithFields(map[string]interface{}{
				"machine": id,
				"tier":    machine.pool.tier,
			}).Changed("Machine removed")
		}
	}

	if len(unrequired) > 0 {
		return pdf(monitor.WithFields(map[string]interface{}{
			"reason":   "unrequire machines replacement",
			"replaced": unrequired,
		}))
	}
	return nil
}
