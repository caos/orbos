package wrap

import (
	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/core"
)

func desire(selfPool string, changesAllowed bool, curr dynamic.Current, svc core.MachinesService, nodeagents map[string]*common.NodeAgentSpec, notifymasters func(machine infra.Machine, peers infra.Machines, vips []*dynamic.VIP) string) func() error {
	return func() error {
		update := []string{selfPool}
	sources:
		for _, source := range curr.Current.SourcePools[selfPool] {
			for _, existing := range update {
				if source == existing {
					continue sources
				}
			}
			update = append(update, source)
		}

		for _, pool := range update {
			if err := curr.Current.Desire(pool, svc, nodeagents, notifymasters); err != nil {
				return err
			}
		}
		return nil
	}
}
