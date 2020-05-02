package wrap

import (
	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/core"
)

var _ core.MachinesService = (*cmpSvcLB)(nil)

type cmpSvcLB struct {
	core.MachinesService
	dynamic       dynamic.Current
	nodeagents    map[string]*common.NodeAgentSpec
	notifymasters string
}

func MachinesService(svc core.MachinesService, curr dynamic.Current, nodeagents map[string]*common.NodeAgentSpec, notifymasters string) *cmpSvcLB {
	return &cmpSvcLB{
		MachinesService: svc,
		dynamic:         curr,
		nodeagents:      nodeagents,
	}
}

func (i *cmpSvcLB) Create(poolName string) (infra.Machine, error) {
	cmp, err := i.MachinesService.Create(poolName)
	if err != nil {
		return nil, err
	}

	desireFunc := desire(poolName, true, i.dynamic, i.MachinesService, i.nodeagents, i.notifymasters)
	return machine(cmp, desireFunc), desireFunc()
}
