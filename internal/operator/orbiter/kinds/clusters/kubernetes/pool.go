package kubernetes

import (
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/kubernetes/edge/k8s"
	"github.com/caos/orbiter/logging"
	"github.com/pkg/errors"
)

/*
type pool struct {
	logger  logging.Logger
	repoURL string
	repoKey string
	desired Pool
	cloud   infra.Pool
	k8s     *k8s.Client
	cmps    []infra.Compute
	mux     sync.Mutex
}

func newPool(
	logger logging.Logger,
	repoURL string,
	repoKey string,
	desired Pool,
	cloudPool infra.Pool,
	k8s *k8s.Client,
	initialComputes []infra.Compute) *pool {
	return &pool{
		logger,
		repoURL,
		repoKey,
		desired,
		cloudPool,
		k8s,
		initialComputes,
		sync.Mutex{},
	}
}*/

// TODO: Implement
/*func (p *pool) deleteComputes(number int) error {

	all := p.computes()
	remaining := all[]
}
*/

func cleanupComputes(logger logging.Logger, pool infra.Pool, k8s *k8s.Client) (err error) {

	nodes, err := k8s.ListNodes()
	if err != nil {
		return err
	}

	computes, err := pool.GetComputes(true)
	if err != nil {
		return err
	}
	logger.WithFields(map[string]interface{}{
		"computes": len(computes),
		"nodes":    len(nodes),
	}).Debug("Aligning computes to nodes")

keepCompute:
	for _, comp := range computes {
		for _, node := range nodes {
			if node.GetName() == comp.ID() {
				continue keepCompute
			}
		}
		if err := comp.Remove(); err != nil {
			return err
		}
	}

	return nil
}

func newComputes(pool infra.Pool, number int) (computes []infra.Compute, err error) {

	computes = make([]infra.Compute, 0)

	var it int
	for it = 0; it < number; it++ {
		var compute infra.Compute
		compute, err = pool.AddCompute()
		if err != nil {
			break
		}
		computes = append(computes, compute)
	}

	if err != nil {
		for _, compute := range computes {
			if rmErr := compute.Remove(); rmErr != nil {
				err = errors.Wrapf(rmErr, "cleaning up compute failed. original error: %s", err)
			}
		}
		return nil, err
	}

	return computes, nil
}
