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
	nodeagents    *common.DesiredNodeAgents
	notifyMasters func(machine infra.Machine, peers infra.Machines, vips []*dynamic.VIP) string
	vrrpInterface string
	vip           func(*dynamic.VIP) string
}

func MachinesService(svc core.MachinesService, curr dynamic.Current, vrrpInterface string, notifyMasters func(machine infra.Machine, peers infra.Machines, vips []*dynamic.VIP) string, vip func(*dynamic.VIP) string) *CmpSvcLB {
	return &CmpSvcLB{
		MachinesService: svc,
		dynamic:         curr,
		notifyMasters:   notifyMasters,
		vrrpInterface:   vrrpInterface,
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
		if !poolDone {
			done = false
		}
		if err != nil {
			return done, err
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
	return c.dynamic.Current.Desire(selfPool, c, c.vrrpInterface, c.notifyMasters, c.vip)
}
