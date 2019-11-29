package adapter

import (
	"github.com/pkg/errors"

	"github.com/caos/infrop/internal/core/operator"
	"github.com/caos/infrop/internal/kinds/clusters/core/infra"
	"github.com/caos/infrop/internal/kinds/providers/static/model"
)

func desireHostname(poolsSpec map[string][]*model.Compute, mapNodeAgent func(cmp infra.Compute) *operator.NodeAgentCurrent) func(compute infra.Compute, pool string) error {
	return func(compute infra.Compute, pool string) error {
		for _, computeSpec := range poolsSpec[pool] {
			if computeSpec.ID == compute.ID() {
				mapNodeAgent(compute).DesireSoftware(&operator.Software{
					Hostname: operator.Package{Config: map[string]string{"hostname": computeSpec.Hostname}},
				})
				return nil
			}
		}
		return errors.Errorf("Compute %s is not configured in pool %s", compute.ID(), pool)
	}
}
