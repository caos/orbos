package wrap

import (
	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"
)

var _ core.MachinesService = (*cmpSvcLB)(nil)

type cmpSvcLB struct {
	core.MachinesService
	dynamic       dynamic.Current
	nodeagents    map[string]*common.NodeAgentSpec
	notifyMasters func(machine infra.Machine, peers infra.Machines, vips []*dynamic.VIP) string
	vrrp          bool
	vip           func(*dynamic.VIP) string
}

func MachinesService(svc core.MachinesService, curr dynamic.Current, nodeagents map[string]*common.NodeAgentSpec, vrrp bool, notifyMasters func(machine infra.Machine, peers infra.Machines, vips []*dynamic.VIP) string, vip func(*dynamic.VIP) string) *cmpSvcLB {
	return &cmpSvcLB{
		MachinesService: svc,
		dynamic:         curr,
		nodeagents:      nodeagents,
		notifyMasters:   notifyMasters,
		vrrp:            vrrp,
		vip:             vip,
	}
}

func (i *cmpSvcLB) Create(poolName string) (infra.Machine, error) {
	cmp, err := i.MachinesService.Create(poolName)
	if err != nil {
		return nil, err
	}

	desireFunc := desire(poolName, i.dynamic, i.MachinesService, i.nodeagents, i.vrrp, i.notifyMasters, i.vip)
	return machine(cmp, desireFunc), desireFunc()
}
