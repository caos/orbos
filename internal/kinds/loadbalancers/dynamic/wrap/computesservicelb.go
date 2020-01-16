package wrap

import (
	"github.com/caos/orbiter/internal/core/operator/common"
	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/kinds/loadbalancers/dynamic"
	"github.com/caos/orbiter/internal/kinds/providers/core"
)

type cmpSvcLB struct {
	original      core.ComputesService
	dynamic       dynamic.Current
	nodeagents    map[string]*common.NodeAgentSpec
	notifymasters string
}

func ComputesService(svc core.ComputesService, curr dynamic.Current, nodeagents map[string]*common.NodeAgentSpec, notifymasters string) core.ComputesService {
	return &cmpSvcLB{
		original:   svc,
		dynamic:    curr,
		nodeagents: nodeagents,
	}
}

func (i *cmpSvcLB) ListPools() ([]string, error) {
	return i.original.ListPools()
}

func (i *cmpSvcLB) List(poolName string, active bool) (infra.Computes, error) {
	return i.original.List(poolName, active)
}

func (i *cmpSvcLB) Create(poolName string) (infra.Compute, error) {
	cmp, err := i.original.Create(poolName)
	if err != nil {
		return nil, err
	}

	desireFunc := desire(poolName, true, i.dynamic, i.original, i.nodeagents, i.notifymasters)
	return compute(cmp, desireFunc), desireFunc()
}
