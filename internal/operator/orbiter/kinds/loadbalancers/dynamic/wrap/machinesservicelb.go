package wrap

import (
	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"
)

var _ core.MachinesService = (*CmpSvcLB)(nil)

type CmpSvcLB struct {
	core.MachinesService
	dynamic       dynamic.Current
	nodeagents    map[string]*common.NodeAgentSpec
	notifyMasters func(machine infra.Machine, peers infra.Machines, vips []*dynamic.VIP) string
	vrrp          bool
	vip           func(*dynamic.VIP) string
}

func MachinesService(svc core.MachinesService, curr dynamic.Current, nodeagents map[string]*common.NodeAgentSpec, vrrp bool, notifyMasters func(machine infra.Machine, peers infra.Machines, vips []*dynamic.VIP) string, vip func(*dynamic.VIP) string) *CmpSvcLB {
	return &CmpSvcLB{
		MachinesService: svc,
		dynamic:         curr,
		nodeagents:      nodeagents,
		notifyMasters:   notifyMasters,
		vrrp:            vrrp,
		vip:             vip,
	}
}

func (i *CmpSvcLB) InitializeDesiredNodeAgents() (bool, error) {
	pools, err := i.ListPools()
	if err != nil {
		return false, err
	}

	done := true
	for _, pool := range pools {
		poolDone, err := i.desire(pool)
		if err != nil {
			return poolDone, err
		}
		if !poolDone {
			done = false
		}
	}
	return done, nil
}

func (i *CmpSvcLB) Create(poolName string) (infra.Machine, error) {
	cmp, err := i.MachinesService.Create(poolName)
	if err != nil {
		return nil, err
	}

	_, err = i.desire(poolName)
	return machine(cmp, func() error {
		_, err := i.desire(poolName)
		return err
	}), err
}

func (c *CmpSvcLB) desire(selfPool string) (bool, error) {
	return c.dynamic.Current.Desire(selfPool, c, c.vrrp, c.notifyMasters, c.vip)
}
