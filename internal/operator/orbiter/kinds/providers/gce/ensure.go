package gce

import (
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	dynamiclbmodel "github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/dynamic/wrap"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/core"
	"github.com/caos/orbiter/internal/push"
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter"
	//	externallbmodel "github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/external"
)

func query(
	desired *Spec,
	current *Current,

	nodeAgentsDesired map[string]*common.NodeAgentSpec,
	lb interface{},

	service core.MachinesService,
) (ensureFunc orbiter.EnsureFunc, err error) {

	current.Current.Ingresses = make(map[string]infra.Address)
	var desireLb func(pool string) error
	switch lbCurrent := lb.(type) {
	case *dynamiclbmodel.Current:

		desireLb = func(pool string) error {
			return lbCurrent.Current.Desire(pool, service, nodeAgentsDesired, "")
		}
		for name, address := range lbCurrent.Current.Addresses {
			current.Current.Ingresses[name] = address
		}
		service = wrap.MachinesService(service, *lbCurrent, nodeAgentsDesired, "")
		//	case *externallbmodel.Current:
		//		for name, address := range lbCurrent.Current.Addresses {
		//			current.Current.Ingresses[name] = address
		//		}
	default:
		return nil, errors.Errorf("Unknown load balancer of type %T", lb)
	}

	pools, err := service.ListPools()
	if err != nil {
		return nil, err
	}

	for _, pool := range pools {
		if err := desireLb(pool); err != nil {
			return nil, err
		}
		if err != nil {
			return nil, err
		}
	}

	return func(_ push.Func) error { return nil }, addPools(current, desired, service)
}
