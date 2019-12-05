package adapter

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"

	"google.golang.org/api/compute/v1"
	"google.golang.org/api/option"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/kinds/providers/core"
	"github.com/caos/orbiter/internal/kinds/providers/edge/ssh"
	"github.com/caos/orbiter/internal/kinds/providers/gce/adapter/concepts"
	"github.com/caos/orbiter/internal/kinds/providers/gce/adapter/resourceservices/backendservice"
	"github.com/caos/orbiter/internal/kinds/providers/gce/adapter/resourceservices/firewallrule"
	"github.com/caos/orbiter/internal/kinds/providers/gce/adapter/resourceservices/forwardingrule"
	"github.com/caos/orbiter/internal/kinds/providers/gce/adapter/resourceservices/healthcheck"
	"github.com/caos/orbiter/internal/kinds/providers/gce/adapter/resourceservices/instance"
	"github.com/caos/orbiter/internal/kinds/providers/gce/adapter/resourceservices/instancegroup"
	"github.com/caos/orbiter/internal/kinds/providers/gce/adapter/resourceservices/targetproxy"
	"github.com/caos/orbiter/internal/kinds/providers/gce/edge/api"
	"github.com/caos/orbiter/internal/kinds/providers/gce/model"
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

func authenticatedService(ctx context.Context, googleApplicationCredentialsValue string) (*compute.Service, error) {
	return compute.NewService(ctx, option.WithCredentialsJSON([]byte(strings.Trim(googleApplicationCredentialsValue, "\""))))
}

func New(logger logging.Logger, id string, lbs map[string]*infra.Ingress, publicKey []byte, privateKeyProperty string) Builder {
	return builderFunc(func(spec model.UserSpec, _ operator.NodeAgentUpdater) (model.Config, Adapter, error) {

		cfg := model.Config{}

		if spec.Verbose && !logger.IsVerbose() {
			logger = logger.Verbose()
		}

		pools := make([]string, 0)
		for poolName := range spec.Pools {
			pools = append(pools, poolName)
		}

		if err := core.Validate(lbs, pools); err != nil {
			return cfg, nil, err
		}

		return cfg, adapterFunc(func(ctx context.Context, secrets *operator.Secrets, deps map[string]interface{}) (*model.Current, error) {

			currentProvider := &infraCurrent{
				pools: make(map[string]infra.Pool),
			}

			current := &model.Current{
				ProviderCurrent: currentProvider,
			}

			creds, err := secrets.Read(id + "_google_application_credentials_value")
			if err != nil {
				return current, err
			}

			svc, err := authenticatedService(ctx, string(creds))
			if err != nil {
				return current, err
			}

			resourceFactory := core.NewResourceFactory(logger, id)

			caller := &api.Caller{Ctx: ctx, OperatorID: id}

			if publicKey == nil {
				suffix := "maintenancekey"
				dynamicKeyProperty := fmt.Sprintf("%s_%s", id, suffix)
				dynamicPubKeyProperty := fmt.Sprintf("%s_%s_pub", id, suffix)

				publicKey, err = ssh.EnsureKeyPair(secrets, dynamicKeyProperty, dynamicPubKeyProperty)
				if err != nil {
					return nil, err
				}
				privateKeyProperty = dynamicKeyProperty
			}

			instancesSvc := instance.NewInstanceService(ctx, logger, id, svc, &spec, caller, secrets, publicKey, privateKeyProperty)

			configuredPools := make([]string, 0)
			for poolName := range spec.Pools {
				configuredPools = append(configuredPools, poolName)
			}

			services := &concepts.Services{
				HealthChecks:   healthcheck.New(logger, svc, &spec, caller),
				BackendService: backendservice.New(logger, svc, &spec, caller),
				InstanceGroup:  instancegroup.New(ctx, logger, svc, &spec, caller),
				ForwardingRule: forwardingrule.New(logger, svc, &spec, caller),
				FirewallRule:   firewallrule.New(logger, svc, &spec, caller),
				TargetProxy:    targetproxy.New(logger, svc, &spec, caller),
			}

			resourcesExecutor := core.NewExecutor([][]core.Cleanupper{
				[]core.Cleanupper{
					services.ForwardingRule,
				},
				[]core.Cleanupper{
					services.TargetProxy,
				},
				[]core.Cleanupper{
					services.BackendService,
				},
				[]core.Cleanupper{
					services.InstanceGroup,
					services.HealthChecks,
				},
			})

			firewallExecutor := core.NewExecutor([][]core.Cleanupper{
				/*		[]core.Cleanupper{ // I guess this is done by the ForwardingRule Cleanupper?
						services.FirewallRule,
					},*/
			})

			var mux sync.RWMutex
			groups := make(map[string][]core.EnsuredGroup)
			currentProvider.ing = make(map[string]infra.Address)

			groupCB := func(poolName string, group core.EnsuredGroup) {
				mux.Lock()
				defer mux.Unlock()
				if _, ok := groups[poolName]; !ok {
					groups[poolName] = make([]core.EnsuredGroup, 0)
				}
				groups[poolName] = append(groups[poolName], group)
			}

			for lbName, lbCfg := range lbs {

				// TODO: Is this correctly scoped? Is it concurrency safe? Does it have to be?
				ipCB := func(ensuredLB string, ip string) {
					mux.Lock()
					defer mux.Unlock()
					currentProvider.ing[ensuredLB] = infra.Address{
						Location: ip,
						Port:     700,
					}
				}

				cfg := &concepts.Config{
					Pools: lbCfg.Pools,
					Ports: []int64{int64(700 /*lbCfg.Port*/)},
					HealthChecks: &healthcheck.Config{
						Path: lbCfg.HealthChecksPath,
						Port: 700, // lbCfg.HealthChecks.Port,
					},
					External: true, //lbCfg.External,
				}
				switch {
				case !cfg.External /*&& (lbCfg.Protocol == model.TCP || lbCfg.Protocol == model.UDP)*/ :
					if err = concepts.PlanInternalTCPUDPLoadBalancing(lbName, resourceFactory, resourcesExecutor, firewallExecutor, services, cfg, ipCB, groupCB); err != nil {
						return current, err
					}
				case cfg.External /*&& (lbCfg.Protocol == model.TCP || lbCfg.Protocol == model.UDP)*/ :
					if err = concepts.PlanTCPProxyLoadBalancing(lbName, resourceFactory, resourcesExecutor, firewallExecutor, services, cfg, ipCB, groupCB); err != nil {
						return current, err
					}
				default:
					err = errors.New("No load balancing concept found for this configuration")
					return current, err
				}
			}

			resourcesCleanupped, err := resourcesExecutor.Run()
			if err != nil {
				return current, err
			}

			instancesSvc.ListPools()

			currentProvider.pools = make(map[string]infra.Pool)
			for poolName := range spec.Pools {
				newPool := core.NewPool(poolName, groups[poolName], instancesSvc)
				if err = newPool.EnsureMembers(); err != nil {
					return current, err
				}
				currentProvider.pools[poolName] = newPool
			}

			existingPools, err := instancesSvc.ListPools()
			if err != nil {
				return current, err
			}

			for _, existingPool := range existingPools {
				if _, ok := currentProvider.pools[existingPool]; !ok {
					currentProvider.pools[existingPool] = core.NewPool(existingPool, make([]core.EnsuredGroup, 1), instancesSvc)
				}
			}
			/*
				for lbName, lbCfg := range loadbalancingCfg {
					if lbCfg.
				}
			*/

			firewallCleanupped, err := firewallExecutor.Run()
			if err != nil {
				return current, err
			}

			cleanupped := make(chan error, 1)
			currentProvider.cu = cleanupped
			go func() {
				var cleanupErr error
				count := 0
				for {
					select {
					case cErr := <-resourcesCleanupped:
						count++
						if cErr != nil {
							cleanupErr = cErr
							continue
						}
						logger.Debug("Resource Cleanupped")
					case cErr := <-firewallCleanupped:
						count++
						if cErr != nil {
							cleanupErr = cErr
							continue
						}
						logger.Debug("Firewalls Cleanupped")
					}
					if count >= 2 {
						if cleanupErr == nil {
							logger.Debug("Cleanupping done")
						}
						cleanupped <- cleanupErr
						return
					}
				}
			}()

			return current, nil
		}), nil
	})
}
