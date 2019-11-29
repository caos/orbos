package core

import "github.com/caos/orbiter/internal/kinds/clusters/core/infra"

type EnsuredGroup interface {
	EnsureMembers(compute []infra.Compute) error
	AddMember(compute infra.Compute) error
}

type ComputesService interface {
	ListPools() ([]string, error)
	List(poolName string, active bool) (infra.Computes, error)
	Create(poolName string) (infra.Compute, error)
}

type pool struct {
	poolName string
	groups   []EnsuredGroup
	svc      ComputesService
}

func NewPool(poolName string, groups []EnsuredGroup, svc ComputesService) infra.Pool {
	return &pool{poolName, groups, svc}
}

func (p *pool) EnsureMembers() error {

	machines, err := p.GetComputes(true)
	if err != nil {
		return err
	}

	for _, group := range p.groups {
		if err := group.EnsureMembers(machines); err != nil {
			return err
		}
	}

	return nil
}

func (p *pool) GetComputes(active bool) (infra.Computes, error) {
	return p.svc.List(p.poolName, active)
}

func (p *pool) AddCompute() (infra.Compute, error) {

	newCompute, err := p.svc.Create(p.poolName)
	if err != nil {
		return nil, err
	}

	for _, group := range p.groups {
		if err := group.AddMember(newCompute); err != nil {
			return nil, err
		}
	}

	return newCompute, nil
}
