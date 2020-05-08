package wrap

import (
	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"
)

type cmpSvcLB struct {
	original      core.MachinesService
	dynamic       dynamic.Current
	nodeagents    map[string]*common.NodeAgentSpec
	notifymasters string
}

func MachinesService(svc core.MachinesService, curr dynamic.Current, nodeagents map[string]*common.NodeAgentSpec, notifymasters string) core.MachinesService {
	return &cmpSvcLB{
		original:   svc,
		dynamic:    curr,
		nodeagents: nodeagents,
	}
}

func (i *cmpSvcLB) ListPools() ([]string, error) {
	return i.original.ListPools()
}

func (i *cmpSvcLB) List(poolName string, active bool) (infra.Machines, error) {
	return i.original.List(poolName, active)
}

func (i *cmpSvcLB) Create(poolName string) (infra.Machine, error) {
	cmp, err := i.original.Create(poolName)
	if err != nil {
		return nil, err
	}

	desireFunc := desire(poolName, true, i.dynamic, i.original, i.nodeagents, i.notifymasters)
	return machine(cmp, desireFunc), desireFunc()
}
