package adapter

import (
	"context"
	"fmt"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/core/operator/orbiter"
	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	dynamiclbmodel "github.com/caos/orbiter/internal/kinds/loadbalancers/dynamic/model"
	"github.com/caos/orbiter/internal/kinds/loadbalancers/dynamic/wrap"
	externallbmodel "github.com/caos/orbiter/internal/kinds/loadbalancers/external/model"
	"github.com/caos/orbiter/internal/kinds/providers/core"
	"github.com/caos/orbiter/internal/kinds/providers/edge/ssh"
	"github.com/caos/orbiter/internal/kinds/providers/static/model"
	"github.com/caos/orbiter/logging"
)

type infraCurrent struct {
	pools map[string]infra.Pool
	ing   map[string]infra.Address
	cu    <-chan error
}

func (i *infraCurrent) Pools() map[string]infra.Pool {
	return i.pools
}

func (i *infraCurrent) Ingresses() map[string]infra.Address {
	return i.ing
}

func (i *infraCurrent) Cleanupped() <-chan error {
	return i.cu
}

func New(logger logging.Logger, id string, healthchecks string, changesDisallowed []string, mapNodeAgent func(cmp infra.Compute) *orbiter.NodeAgentCurrent) Builder {
	return builderFunc(func(spec model.UserSpec, _ orbiter.NodeAgentUpdater) (model.Config, Adapter, error) {

		cfg := model.Config{
			Logger:       logger,
			ID:           id,
			Healthchecks: healthchecks,
		}

		if spec.Verbose && !logger.IsVerbose() {
			logger = logger.Verbose()
		}

		return cfg, adapterFunc(func(ctx context.Context, secrets *orbiter.Secrets, deps map[string]interface{}) (*model.Current, error) {

			currentProvider := &infraCurrent{
				pools: make(map[string]infra.Pool),
				ing:   make(map[string]infra.Address),
			}

			current := &model.Current{
				ProviderCurrent: currentProvider,
			}

			bootstrapKeyProperty := fmt.Sprintf("%s_bootstrapkey", id)

			suffix := "maintenancekey"
			dynamicKeyProperty := fmt.Sprintf("%s_%s", id, suffix)
			dynamicPubKeyProperty := fmt.Sprintf("%s_%s_pub", id, suffix)

			publicKey, err := ssh.EnsureKeyPair(secrets, dynamicKeyProperty, dynamicPubKeyProperty)
			if err != nil {
				return nil, err
			}

			// TODO: Allow Changes
			desireHostnameFunc := desireHostname(spec.Pools, mapNodeAgent)

			computesSvc := NewComputesService(logger, id, &spec, bootstrapKeyProperty, dynamicKeyProperty, publicKey, secrets, desireHostnameFunc)
			pools, err := computesSvc.ListPools()
			if err != nil {
				return nil, err
			}
			for _, pool := range pools {
				computes, err := computesSvc.List(pool, true)
				if err != nil {
					return nil, err
				}
				for _, compute := range computes {
					if err := desireHostnameFunc(compute, pool); err != nil {
						return nil, err
					}
				}
			}

			for depName, dep := range deps {
				switch lb := dep.(type) {
				case *dynamiclbmodel.Current:
					for name, address := range lb.Addresses {
						currentProvider.ing[name] = address
					}
					for _, pool := range pools {
						changesAllowed := true
						for _, disallowed := range changesDisallowed {
							if pool == disallowed {
								changesAllowed = false
							}
						}
						if err := lb.Desire(pool, changesAllowed, computesSvc, mapNodeAgent, ""); err != nil {
							return nil, err
						}
					}
					computesSvc = wrap.ComputesService(computesSvc, *lb, mapNodeAgent, "")
				case *externallbmodel.Current:
					currentProvider.ing[depName] = lb.Address
				default:
					return nil, errors.Errorf("Unknown load balancer of type %T", lb)
				}
			}

			currentProvider.pools = make(map[string]infra.Pool)
			for pool := range spec.Pools {
				currentProvider.pools[pool] = core.NewPool(pool, nil, computesSvc)
			}

			unconfiguredPools, err := computesSvc.ListPools()
			if err != nil {
				return current, nil
			}
			for _, unconfiguredPool := range unconfiguredPools {
				if _, ok := currentProvider.pools[unconfiguredPool]; !ok {
					currentProvider.pools[unconfiguredPool] = core.NewPool(unconfiguredPool, nil, computesSvc)
				}
			}

			cu := make(chan error)
			go func() {
				cu <- nil
			}()
			currentProvider.cu = cu
			return current, nil
		}), nil
	})
}
