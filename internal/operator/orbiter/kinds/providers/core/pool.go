package core

import "github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"

// TODO: Do we still need this?
type MachinesService interface {
	ListPools() ([]string, error)
	List(poolName string) (infra.Machines, error)
	Create(poolName string) (infra.Machine, error)
}
