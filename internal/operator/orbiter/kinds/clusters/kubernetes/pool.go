package kubernetes

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

/*
type pool struct {
	monitor  mntr.Monitor
	repoURL string
	repoKey string
	desired Pool
	cloud   infra.Pool
	k8s     *k8s.Client
	cmps    []infra.Machine
	mux     sync.Mutex
}

func newPool(
	monitor mntr.Monitor,
	repoURL string,
	repoKey string,
	desired Pool,
	cloudPool infra.Pool,
	k8s *k8s.Client,
	initialMachines []infra.Machine) *pool {
	return &pool{
		monitor,
		repoURL,
		repoKey,
		desired,
		cloudPool,
		k8s,
		initialMachines,
		sync.Mutex{},
	}
}*/

// TODO: Implement
/*func (p *pool) deleteMachines(number int) error {

	all := p.machines()
	remaining := all[]
}
*/

func cleanupMachines(monitor mntr.Monitor, pool infra.Pool, k8s *Client) (err error) {

	nodes, err := k8s.ListNodes()
	if err != nil {
		return err
	}

	machines, err := pool.GetMachines()
	if err != nil {
		return err
	}
	monitor.WithFields(map[string]interface{}{
		"machines": len(machines),
		"nodes":    len(nodes),
	}).Debug("Aligning machines to nodes")

keepMachine:
	for _, comp := range machines {
		for _, node := range nodes {
			if node.GetName() == comp.ID() {
				continue keepMachine
			}
		}
		if err := comp.Remove(); err != nil {
			return err
		}
	}

	return nil
}

func newMachines(pool infra.Pool, number int) (machines []infra.Machine, err error) {

	machines = make([]infra.Machine, 0)

	var it int
	for it = 0; it < number; it++ {
		var machine infra.Machine
		machine, err = pool.AddMachine()
		if err != nil {
			break
		}
		machines = append(machines, machine)
	}

	if err != nil {
		for _, machine := range machines {
			if rmErr := machine.Remove(); rmErr != nil {
				err = errors.Wrapf(rmErr, "cleaning up machine failed. original error: %s", err)
			}
		}
		return nil, err
	}

	return machines, nil
}
