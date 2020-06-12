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
			InternalPort: internalPort(lb),
			ExternalPort: externalPort(lb),
		}
	}

	desireHealthcheck := func(pool string, machine infra.Machine) {
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
				}
			}
		}
	}
	context.machinesService.onCreate = desireHealthcheck

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

		return ensureLB()

	}, initPools(current, desired, context, normalized, lbCurrent, nodeAgentsDesired)
}

func internalPort(lb *normalizedLoadbalancer) uint16 {
	return lb.port
}

func externalPort(lb *normalizedLoadbalancer) uint16 {
	port, err := strconv.Atoi(strings.Split(lb.forwardingRule.gce.PortRange, "-")[0])
	if err != nil {
		panic(err)
	}
	return uint16(port)
}
