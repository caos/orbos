package gce

import (
	"fmt"
	"github.com/caos/orbos/internal/api"
	"strconv"
	"strings"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic/wrap"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	dynamiclbmodel "github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/pkg/errors"

	"github.com/caos/orbos/internal/operator/orbiter"
	//	externallbmodel "github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/external"
)

func query(
	desired *Spec,
	current *Current,
	lb interface{},
	context *context,
	currentNodeAgents map[string]*common.NodeAgentCurrent,
	nodeAgentsDesired map[string]*common.NodeAgentSpec,
	orbiterCommit,
	repoURL,
	repoKey string,
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

	queryNA, installNA := core.NodeAgentFuncs(context.monitor, orbiterCommit, repoURL, repoKey, currentNodeAgents)

	desireNodeAgent := func(pool string, machine infra.Machine) error {

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
		running, err := queryNA(machine)
		if err != nil {
			return err
		}
		if !running {
			return installNA(machine)
		}

		return nil
	}

	context.machinesService.onCreate = desireNodeAgent
	wrappedMachines := wrap.MachinesService(context.machinesService, *lbCurrent, nodeAgentsDesired, false, nil, func(vip *dynamic.VIP) string {
		for _, transport := range vip.Transport {
			address, ok := current.Current.Ingresses[transport.Name]
			if ok {
				return address.Location
			}
		}
		panic(fmt.Errorf("external address for %v is not ensured", vip))
	})
	return func(psf api.SecretFunc) *orbiter.EnsureResult {

		if err := ensureGcloud(context); err != nil {
			return orbiter.ToEnsureResult(false, err)
		}

		if err := ensureIdentityAwareProxyAPIEnabled(context); err != nil {
			return orbiter.ToEnsureResult(false, err)
		}

		if err := ensureCloudNAT(context); err != nil {
			return orbiter.ToEnsureResult(false, err)
		}

		if err := context.machinesService.restartPreemptibleMachines(); err != nil {
			return orbiter.ToEnsureResult(false, err)
		}

		pools, err := context.machinesService.ListPools()
		if err != nil {
			return orbiter.ToEnsureResult(false, err)
		}

		if err := ensureLB(); err != nil {
			return orbiter.ToEnsureResult(false, err)
		}

		for _, pool := range pools {
			machines, err := context.machinesService.List(pool)
			if err != nil {
				return orbiter.ToEnsureResult(false, err)
			}
			for _, machine := range machines {
				if err := desireNodeAgent(pool, machine); err != nil {
					return orbiter.ToEnsureResult(false, err)
				}
			}
		}

		return orbiter.ToEnsureResult(wrappedMachines.InitializeDesiredNodeAgents())
	}, initPools(current, desired, context, normalized, wrappedMachines)
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
