package static

import (
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
)

func desireHostname(poolsSpec map[string][]*Compute, nodeagents map[string]*common.NodeAgentSpec) func(compute infra.Compute, pool string) error {
	return func(compute infra.Compute, pool string) error {
		for _, computeSpec := range poolsSpec[pool] {
			if computeSpec.ID == compute.ID() {
				nodeagent, ok := nodeagents[computeSpec.ID]
				if !ok {
					nodeagent = &common.NodeAgentSpec{}
					nodeagents[computeSpec.ID] = nodeagent
				}
				if nodeagent.Software == nil {
					nodeagent.Software = &common.Software{}
				}
				nodeagent.Software.Hostname = common.Package{Config: map[string]string{"hostname": computeSpec.Hostname}}
				return nil
			}
		}
		return errors.Errorf("Compute %s is not configured in pool %s", compute.ID(), pool)
	}
}
