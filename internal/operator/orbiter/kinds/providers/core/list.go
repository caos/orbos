package core

import (
	"sync"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
)

func ListMachines(svc MachinesService) (map[string]infra.Machine, error) {
	machines := make(map[string]infra.Machine, 0)
	var mux sync.Mutex
	return machines, Each(svc, func(pool string, machine infra.Machine) error {
		mux.Lock()
		defer mux.Unlock()
		machines[machine.ID()] = machine
		return nil
	})
}
