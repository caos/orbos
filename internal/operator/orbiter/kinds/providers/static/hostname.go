package static

import (
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
)

func desireHostname(poolsSpec map[string][]*Machine, nodeagents map[string]*common.NodeAgentSpec) func(machine infra.Machine, pool string) error {
	return func(machine infra.Machine, pool string) error {
		for _, machineSpec := range poolsSpec[pool] {
			if machineSpec.ID == machine.ID() {
				nodeagent, ok := nodeagents[machineSpec.ID]
				if !ok {
					nodeagent = &common.NodeAgentSpec{}
					nodeagents[machineSpec.ID] = nodeagent
				}
				if nodeagent.Software == nil {
					nodeagent.Software = &common.Software{}
				}
				nodeagent.Software.Hostname = common.Package{Config: map[string]string{"hostname": machineSpec.Hostname}}
				return nil
			}
		}
		return errors.Errorf("Machine %s is not configured in pool %s", machine.ID(), pool)
	}
}
