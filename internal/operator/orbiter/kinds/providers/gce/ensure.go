package gce

import (
	"fmt"
	"strings"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	dynamiclbmodel "github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/ssh"
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
	current.Current.Ingresses = make(map[string]*infra.Address)
	normalized := normalize(context.monitor, lbCurrent.Current.Spec, context.orbID, context.providerID)

	ensureLB, err := queryResources(context, normalized)
	if err != nil {
		return nil, err
	}

	for _, lb := range normalized {
		current.Current.Ingresses[lb.transport] = &infra.Address{
			Location: lb.address.gce.Address,
			Port:     uint16(lb.healthcheck.gce.Port),
			Bind: func(_ string) string {
				return "0.0.0.0"
			},
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
		for _, lb := range normalized {
			for _, destPool := range lb.targetPool.destPools {
				if pool == destPool {
					na.Software.Health.Config[fmt.Sprintf(
						"%s:%d%s",
						lb.healthcheck.gce.Host,
						lb.healthcheck.gce.Port,
						lb.healthcheck.gce.RequestPath)] = fmt.Sprintf(
						"%d@%s://%s:%d%s",
						lb.healthcheck.desired.Code,
						lb.healthcheck.desired.Protocol,
						machine.IP(),
						strings.Split(lb.forwardingRule.gce.PortRange, "-")[0],
						lb.healthcheck.desired.Path,
					)
				}
			}
		}
	}
	context.machinesService.onCreate = desireHealthcheck

	return func(psf push.Func) error {

		if err := ensureLB(); err != nil {
			return err
		}

		if desired.SSHKey != nil && desired.SSHKey.Private != nil && desired.SSHKey.Private.Value != "" && desired.SSHKey.Public != nil && desired.SSHKey.Public.Value != "" {
			return nil
		}
		private, public, err := ssh.Generate()
		if err != nil {
			return err
		}
		desired.SSHKey.Private.Value = private
		desired.SSHKey.Public.Value = public
		if err := psf(context.monitor.WithField("secret", "sshkey")); err != nil {
			return err
		}

		return nil
	}, initPools(current, desired, context.machinesService)
}
