package static

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"
)

func destroy(svc *machinesService, desired *DesiredV0, current *Current) error {

	core.Each(svc, func(pool string, machine infra.Machine) error {
		return machine.Remove()
	})

	return addPools(current, desired, svc)
}
