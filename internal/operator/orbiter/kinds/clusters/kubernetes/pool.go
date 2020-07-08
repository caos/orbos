package kubernetes

import (
	"sync"

	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/mntr"
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

	var wg sync.WaitGroup
	var it int
	for it = 0; it < number; it++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			machine, addErr := pool.AddMachine()
			if addErr != nil {
				err = helpers.Concat(err, addErr)
				return
			}
			machines = append(machines, machine)
		}()
	}

	wg.Wait()

	if err != nil {
		for _, machine := range machines {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err = helpers.Concat(err, machine.Remove())
			}()
		}
		wg.Wait()
	}

	return machines, err
}
