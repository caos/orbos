package static

import (
	"fmt"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/mntr"
)

func desireHostname(poolsSpec map[string][]*Machine, nodeagents *common.DesiredNodeAgents, nodeagentsCurr *common.CurrentNodeAgents, monitor mntr.Monitor) func(machine infra.Machine, pool string) (bool, error) {
	return func(machine infra.Machine, pool string) (bool, error) {
		for _, machineSpec := range poolsSpec[pool] {
			if machineSpec.ID == machine.ID() {
				machineMonitor := monitor.WithFields(map[string]interface{}{
					"machine": machine.ID(),
				})

				nodeagent, _ := nodeagents.Get(machineSpec.ID)
				if nodeagent.Software.Hostname.Config == nil || nodeagent.Software.Hostname.Config["hostname"] != machineSpec.ID {
					nodeagent.Software.Hostname = common.Package{Config: map[string]string{"hostname": machineSpec.ID}}
					machineMonitor.Changed("Hostname desired")
				}
				logWaiting := func() {
					machineMonitor.Info("Awaiting hostname")
				}
				curr, ok := nodeagentsCurr.Get(machine.ID())
				if !ok {
					logWaiting()
					return false, nil
				}
				if curr.Software.Hostname.Config == nil || curr.Software.Hostname.Config["hostname"] != machineSpec.ID {
					logWaiting()
					return false, nil
				}
				return true, nil
			}
		}
		return false, fmt.Errorf("machine %s is not configured in pool %s", machine.ID(), pool)
	}
}
