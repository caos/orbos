package kubernetes

import (
	"time"

	"github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/mntr"
)

func maintainNodes(allInitializedMachines initializedMachines, monitor mntr.Monitor, k8sClient *Client, pdf api.PushDesiredFunc) (done bool, err error) {

	allInitializedMachines.forEach(monitor, func(machine *initializedMachine, machineMonitor mntr.Monitor) bool {
		if err = machine.reconcile(); err != nil {
			return false
		}
		return true
	})
	if err != nil {
		return false, err
	}

	allInitializedMachines.forEach(monitor, func(machine *initializedMachine, machineMonitor mntr.Monitor) bool {
		req, unreq := machine.infra.RebootRequired()
		if !req {
			return true
		}
		if err = k8sClient.Drain(machine.currentMachine, machine.node, rebooting); err != nil {
			return false
		}
		machine.currentMachine.Rebooting = true
		machineMonitor.Info("Requiring reboot")
		unreq()
		machine.desiredNodeagent.RebootRequired = time.Now().Truncate(time.Minute)
		err = pdf(monitor.WithField("reason", "remove machine from reboot list"))
		return false
	})
	if err != nil {
		return false, err
	}

	done = true
	allInitializedMachines.forEach(monitor, func(machine *initializedMachine, machineMonitor mntr.Monitor) bool {
		if !machine.currentMachine.FirewallIsReady {
			done = false
			machineMonitor.Info("Node agent is not ready yet")
		}
		return true
	})
	return done, nil
}
