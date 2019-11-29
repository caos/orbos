package adapter

import (
	"sync"

	"github.com/pkg/errors"

	"github.com/caos/infrop/internal/kinds/clusters/kubernetes/edge/k8s"
	"github.com/caos/infrop/internal/kinds/clusters/kubernetes/model"

	"github.com/caos/infrop/internal/core/helpers"
	"github.com/caos/infrop/internal/kinds/clusters/core/infra"
)

type pool struct {
	cfg               *model.Config
	nodeagentFullPath func(compute infra.Compute) []string
	poolSpec          *poolSpec
	cloud             infra.Pool
	k8s               *k8s.Client
	cmps              []infra.Compute
	mux               sync.Mutex
}

type poolSpec struct {
	group string
	spec  *model.Pool
}

func newPool(
	cfg *model.Config,
	nodeagentFullPath func(compute infra.Compute) []string,
	poolSpec *poolSpec,
	cloudPool infra.Pool,
	k8s *k8s.Client,
	initialComputes []infra.Compute) *pool {
	return &pool{
		cfg,
		nodeagentFullPath,
		poolSpec,
		cloudPool,
		k8s,
		initialComputes,
		sync.Mutex{},
	}
}

func (p *pool) computes() []infra.Compute {
	return p.cmps
}

// TODO: Implement
/*func (p *pool) deleteComputes(number int) error {

	all := p.computes()
	remaining := all[]
}
*/

func (p *pool) cleanupComputes() (err error) {

	remainingComputes := make([]infra.Compute, 0)
	nodes, err := p.k8s.ListNodes()
	if err != nil {
		return err
	}

	p.cfg.Params.Logger.WithFields(map[string]interface{}{
		"computes": len(p.cmps),
		"nodes":    len(nodes),
	}).Debug("Aligning computes to nodes")

	var wg sync.WaitGroup
	synchronizer := helpers.NewSynchronizer(&wg)

keepCompute:
	for _, comp := range p.cmps {
		for _, node := range nodes {
			if node.GetName() == comp.ID() {
				remainingComputes = append(remainingComputes, comp)
				continue keepCompute
			}
		}
		wg.Add(1)
		go func(compute infra.Compute) {
			synchronizer.Done(compute.Remove())
		}(comp)
	}

	wg.Wait()
	if synchronizer.IsError() {
		return synchronizer
	}

	p.mux.Lock()
	p.cmps = remainingComputes
	p.mux.Unlock()

	return nil
}

func (p *pool) newComputes(number int, callback func(infra.Compute)) (err error) {

	defer func() {
		if err != nil {
			p.cfg.Params.Logger.WithFields(map[string]interface{}{
				"message": err.Error(),
			}).Debug("New computes retured error")
		}
	}()

	var wg sync.WaitGroup
	synchronizer := helpers.NewSynchronizer(&wg)
	for it := 0; it < number; it++ {
		wg.Add(1)
		go func() {
			synchronizer.Done(p.newCompute(callback))
		}()
	}
	wg.Wait()
	if synchronizer.IsError() {
		return errors.Wrap(synchronizer, "creating new computes failed")
	}
	return nil
}

func (p *pool) newCompute(callback func(infra.Compute)) (err error) {

	p.mux.Lock()
	defer p.mux.Unlock()
	compute, err := p.cloud.AddCompute()
	if err != nil {
		return err
	}
	p.mux.Unlock()

	//TODO: Remove. Instead try again later
	defer func() {
		if err != nil {
			err = errors.Wrapf(err, "creating new compute %s failed", compute.ID())
			if rmErr := compute.Remove(); rmErr != nil {
				panic(errors.Errorf("Error cleaning up after error creating new compute. Please remove compute %s manually. Original error: %s. Cleanup error: %s", compute.ID(), err.Error(), rmErr.Error()))
			}
		}
	}()

	if err := installNodeAgent(p.cfg, compute, p.nodeagentFullPath); err != nil {
		return err
	}

	p.mux.Lock()
	p.cmps = append(p.cmps, compute)
	callback(compute)
	return nil
}
