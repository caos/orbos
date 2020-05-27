package core

import "github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"

/*
type EnsuredGroup interface {
	EnsureMembers(machine []infra.Machine) error
	AddMember(machine infra.Machine) error
}
*/
// TODO: Do we still need this?
type MachinesService interface {
	ListPools() ([]string, error)
	List(poolName string) (infra.Machines, error)
	Create(poolName string) (infra.Machine, error)
}

/*
type pool struct {
	poolName string
	groups   []EnsuredGroup
	svc      MachinesService
}

func NewPool(poolName string, groups []EnsuredGroup, svc MachinesService) infra.Pool {
	return &pool{poolName, groups, svc}
}

func (p *pool) EnsureMembers() error {

	machines, err := p.GetMachines()
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

func (p *pool) GetMachines() (infra.Machines, error) {
	return p.svc.List(p.poolName)
}

func (p *pool) AddMachine() (infra.Machine, error) {

	newMachine, err :=
	return p.svc.Create(p.poolName)
		if err != nil {
			return nil, err
		}
		for _, group := range p.groups {
			if err := group.AddMember(newMachine); err != nil {
				return nil, err
			}
		}

		return newMachine, nil
}
*/
