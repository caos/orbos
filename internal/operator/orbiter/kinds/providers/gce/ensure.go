package gce

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	dynamiclbmodel "github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbos/internal/push"
	"github.com/pkg/errors"

	"github.com/caos/orbos/internal/operator/orbiter"
	//	externallbmodel "github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/external"
)

func query(
	desired *Spec,
	current *Current,
	lb interface{},
	context *context,
	nodeAgentsDesired map[string]*common.NodeAgentSpec,
) (ensureFunc orbiter.EnsureFunc, err error) {

	lbCurrent, ok := lb.(*dynamiclbmodel.Current)
	if !ok {
		errors.Errorf("Unknown or unsupported load balancing of type %T", lb)
	}
	normalized, firewalls := normalize(context.monitor, lbCurrent.Current.Spec, context.orbID, context.providerID)

	ensureLB, err := queryResources(context, normalized, firewalls)
	if err != nil {
		return nil, err
	}

	current.Current.Ingresses = make(map[string]*infra.Address)
	for _, lb := range normalized {
		current.Current.Ingresses[lb.transport] = &infra.Address{
			Location:     lb.address.gce.Address,
			BackendPort:  internalPort(lb),
			FrontendPort: externalPort(lb),
		}
	}

	desireNodeAgent := func(pool string, machine infra.Machine) {
		machineID := machine.ID()
		na, ok := nodeAgentsDesired[machineID]
		if !ok {
			na = &common.NodeAgentSpec{}
			nodeAgentsDesired[machineID] = na
		}
		if na.Software == nil {
			na.Software = &common.Software{}
		}
		if na.Software.Health.Config == nil {
			na.Software.Health.Config = make(map[string]string)
		}
		if na.Firewall == nil {
			fw := common.Firewall(make(map[string]*common.Allowed))
			na.Firewall = &fw
		}
		lbCurrent.Current.Desire(pool, context.machinesService, nodeAgentsDesired, false, nil, func(vip *dynamiclbmodel.VIP) string {
			for _, transport := range vip.Transport {
				for _, lb := range normalized {
					if lb.transport == transport.Name {
						return lb.address.gce.Address
					}
				}
			}
			return "UNKNOWN"
		})

		for _, lb := range normalized {
			for _, destPool := range lb.targetPool.destPools {
				if pool == destPool {
					na.Software.Health.Config[fmt.Sprintf(
						"%s:%d%s",
						"0.0.0.0",
						lb.healthcheck.gce.Port,
						lb.healthcheck.gce.RequestPath)] = fmt.Sprintf(
						"%d@%s://%s:%d%s",
						lb.healthcheck.desired.Code,
						lb.healthcheck.desired.Protocol,
						machine.IP(),
						internalPort(lb),
						lb.healthcheck.desired.Path,
					)
					na.Firewall.Merge(map[string]*common.Allowed{
						lb.healthcheck.gce.Description: {
							Port:     fmt.Sprintf("%d", lb.healthcheck.gce.Port),
							Protocol: "tcp",
						},
					})
					break
				}
			}
		}
	}
	context.machinesService.onCreate = desireNodeAgent

	return func(psf push.Func) error {

		if err := ensureGcloud(context); err != nil {
			return err
		}

		if err := ensureIdentityAwareProxyAPIEnabled(context); err != nil {
			return err
		}

		if err := ensureCloudNAT(context); err != nil {
			return err
		}

		if err := context.machinesService.restartPreemptibleMachines(); err != nil {
			return err
		}

		pools, err := context.machinesService.ListPools()
		if err != nil {
			return err
		}

		if err := ensureLB(); err != nil {
			return err
		}

		for _, pool := range pools {
			machines, err := context.machinesService.List(pool)
			if err != nil {
				return err
			}
			for _, machine := range machines {
				desireNodeAgent(pool, machine)
			}
		}
		return nil
	}, initPools(current, desired, context, normalized, lbCurrent, nodeAgentsDesired)
}

func internalPort(lb *normalizedLoadbalancer) uint16 {
	return lb.backendPort
}

func externalPort(lb *normalizedLoadbalancer) uint16 {
	port, err := strconv.Atoi(strings.Split(lb.forwardingRule.gce.PortRange, "-")[0])
	if err != nil {
		panic(err)
	}
	return uint16(port)
}
