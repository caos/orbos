package static

import (
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
)

func desireHostname(poolsSpec map[string][]*Machine, nodeagents map[string]*common.NodeAgentSpec, nodeagentsCurr map[string]*common.NodeAgentCurrent, monitor mntr.Monitor) func(machine infra.Machine, pool string) (bool, error) {
	return func(machine infra.Machine, pool string) (bool, error) {
		for _, machineSpec := range poolsSpec[pool] {
			if machineSpec.ID == machine.ID() {
				nodeagent, ok := nodeagents[machineSpec.ID]
				machineMonitor := monitor.WithFields(map[string]interface{}{
					"machine":  machine.ID(),
					"hostname": machineSpec.Hostname,
				})
				if !ok {
					nodeagent = &common.NodeAgentSpec{}
					nodeagents[machineSpec.ID] = nodeagent
				}
				if nodeagent.Software == nil {
					nodeagent.Software = &common.Software{}
				}

				if nodeagent.Software.Hostname.Config == nil || nodeagent.Software.Hostname.Config["hostname"] != machineSpec.Hostname {
					nodeagent.Software.Hostname = common.Package{Config: map[string]string{"hostname": machineSpec.Hostname}}
					machineMonitor.Changed("Hostname desired")
				}
				logWaiting := func() {
					machineMonitor.Info("Awaiting hostname")
				}
				if nodeagentsCurr == nil {
					logWaiting()
					return false, nil
				}
				curr, ok := nodeagentsCurr[machine.ID()]
				if !ok || curr == nil {
					logWaiting()
					return false, nil
				}
				if curr.Software.Hostname.Config == nil || curr.Software.Hostname.Config["hostname"] != machineSpec.Hostname {
					logWaiting()
					return false, nil
				}
				return true, nil
			}
		}
		return false, errors.Errorf("Machine %s is not configured in pool %s", machine.ID(), pool)
	}
}
