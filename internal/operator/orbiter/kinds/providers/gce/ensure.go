package gce

import (
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	dynamiclbmodel "github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/ssh"
	"github.com/caos/orbiter/internal/push"
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/orbiter"
	//	externallbmodel "github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/external"
)

func query(
	desired *Spec,
	current *Current,
	lb interface{},
	context *context,
) (ensureFunc orbiter.EnsureFunc, err error) {

	lbCurrent, ok := lb.(*dynamiclbmodel.Current)
	if !ok {
		errors.Errorf("Unknown or unsupported load balancing of type %T", lb)
	}
	current.Current.Ingresses = make(map[string]*infra.Address)
	normalized := normalize(context.monitor, lbCurrent.Current.Spec, context.orbID, context.providerID)
	for _, lb := range normalized {
		current.Current.Ingresses[lb.transport] = &infra.Address{
			Location: lb.address.gce.Address,
			Port:     uint16(lb.healthcheck.gce.Port),
		}
	}

	return func(psf push.Func) error {

		if err := chain(
			context, normalized,
			queryHealthchecks,
			queryTargetPools,
			queryAddresses,
			queryForwardingRules,
			queryFirewall,
		); err != nil {
			return err
		}

		for _, pool := range lbCurrent.Current.Spec {
			for _, vip := range pool {
				for _, src := range vip.Transport {
					current.Current.Ingresses[src.Name] = &infra.Address{
						Location: vip.IP,
						Port:     uint16(src.SourcePort),
					}
				}
			}
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
	}, addPools(current, desired, context.machinesService)
}
