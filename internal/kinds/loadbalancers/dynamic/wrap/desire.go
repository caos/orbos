package wrap

import (
	"github.com/caos/infrop/internal/core/operator"
	"github.com/caos/infrop/internal/kinds/clusters/core/infra"
	"github.com/caos/infrop/internal/kinds/loadbalancers/dynamic/model"
	"github.com/caos/infrop/internal/kinds/providers/core"
)

func desire(selfPool string, changesAllowed bool, dynamic model.Current, svc core.ComputesService, nodeagent func(infra.Compute) *operator.NodeAgentCurrent) func() error {
	return func() error {
		update := []string{selfPool}
	sources:
		for _, source := range dynamic.SourcePools[selfPool] {
			for _, existing := range update {
				if source == existing {
					continue sources
				}
			}
			update = append(update, source)
		}

		for _, pool := range update {
			if err := dynamic.Desire(pool, changesAllowed, svc, nodeagent); err != nil {
				return err
			}
		}
		return nil
	}
}
