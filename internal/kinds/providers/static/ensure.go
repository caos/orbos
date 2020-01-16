package static

import (
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/core/operator/orbiter"
"github.com/caos/orbiter/internal/core/operator/common"
	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	dynamiclbmodel "github.com/caos/orbiter/internal/kinds/loadbalancers/dynamic"
	"github.com/caos/orbiter/internal/kinds/loadbalancers/dynamic/wrap"

	//	externallbmodel "github.com/caos/orbiter/internal/kinds/loadbalancers/external"
	"github.com/caos/orbiter/internal/kinds/providers/core"
	"github.com/caos/orbiter/internal/kinds/providers/edge/ssh"
	"github.com/caos/orbiter/logging"
)

func ensure(
	desired *DesiredV0,
	current *Current,
	sec *SecretsV0,

	nodeAgentsDesired map[string]*common.NodeAgentSpec,
	lb interface{},
	masterkey string,

	logger logging.Logger,
	id string,
) (err error) {

	if sec.Secrets.Maintenance.Private == nil && sec.Secrets.Maintenance.Public == nil {
		priv, pub, err := ssh.Generate()
		if err != nil {
			return err
		}
		sec.Secrets.Maintenance.Private = &orbiter.Secret{Value: priv}
		sec.Secrets.Maintenance.Public = &orbiter.Secret{Value: pub}
	}

	// TODO: Allow Changes
	desireHostnameFunc := desireHostname(desired.Spec.Pools, nodeAgentsDesired)

	computesSvc := NewComputesService(logger, desired, []byte(sec.Secrets.Bootstrap.Private.Value), []byte(sec.Secrets.Maintenance.Private.Value), []byte(sec.Secrets.Maintenance.Public.Value), id, desireHostnameFunc)
	pools, err := computesSvc.ListPools()
	if err != nil {
		return err
	}
	for _, pool := range pools {
		computes, err := computesSvc.List(pool, true)
		if err != nil {
			return err
		}
		for _, compute := range computes {
			if err := desireHostnameFunc(compute, pool); err != nil {
				return err
			}
		}
	}

	current.Current.Ingresses = make(map[string]infra.Address)
	switch lbCurrent := lb.(type) {
	case *dynamiclbmodel.Current:
		for name, address := range lbCurrent.Current.Addresses {
			current.Current.Ingresses[name] = address
		}
		for _, pool := range pools {
			if err := lbCurrent.Current.Desire(pool, computesSvc, nodeAgentsDesired, ""); err != nil {
				return err
			}
		}
		computesSvc = wrap.ComputesService(computesSvc, *lbCurrent, nodeAgentsDesired, "")
		//	case *externallbmodel.Current:
		//		for name, address := range lbCurrent.Current.Addresses {
		//			current.Current.Ingresses[name] = address
		//		}
	default:
		return errors.Errorf("Unknown load balancer of type %T", lb)
	}

	current.Current.Pools = make(map[string]infra.Pool)
	for pool := range desired.Spec.Pools {
		current.Current.Pools[pool] = core.NewPool(pool, nil, computesSvc)
	}

	unconfiguredPools, err := computesSvc.ListPools()
	if err != nil {
		return nil
	}
	for _, unconfiguredPool := range unconfiguredPools {
		if _, ok := current.Current.Pools[unconfiguredPool]; !ok {
			current.Current.Pools[unconfiguredPool] = core.NewPool(unconfiguredPool, nil, computesSvc)
		}
	}

	cu := make(chan error)
	go func() {
		cu <- nil
	}()
	current.Current.cleanupped = cu
	return nil
}
