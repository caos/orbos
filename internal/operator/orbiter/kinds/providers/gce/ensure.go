package gce

import (
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	dynamiclbmodel "github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/dynamic/wrap"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/core"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/ssh"
	"github.com/caos/orbiter/internal/push"
	"github.com/caos/orbiter/mntr"
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter"
	//	externallbmodel "github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/external"
)

func query(
	desired *Spec,
	current *Current,
	monitor mntr.Monitor,

	nodeAgentsDesired map[string]*common.NodeAgentSpec,
	lb interface{},

	machinesSvc core.MachinesService,
	addressesSvc *addressesSvc,
) (ensureFunc orbiter.EnsureFunc, err error) {

	current.Current.Ingresses = make(map[string]*infra.Address)
	var desireLb func(pool string) error

	var addresses map[string]*infra.Address

	switch lbCurrent := lb.(type) {
	case *dynamiclbmodel.Current:
		addresses = lbCurrent.Current.Addresses
		desireLb = func(pool string) error {
			return lbCurrent.Current.Desire(pool, machinesSvc, nodeAgentsDesired, "")
		}
		for name, address := range lbCurrent.Current.Addresses {
			current.Current.Ingresses[name] = address
		}
		machinesSvc = wrap.MachinesService(machinesSvc, *lbCurrent, nodeAgentsDesired, "")
		//	case *externallbmodel.Current:
		//		for name, address := range lbCurrent.Current.Addresses {
		//			current.Current.Ingresses[name] = address
		//		}
	default:
		return nil, errors.Errorf("Unknown load balancer of type %T", lb)
	}

	return func(psf push.Func) error {
		changed, err := addressesSvc.ensure(addresses)
		if err != nil {
			return err
		}

		if changed {
			if err := psf(monitor.WithField("ips", addresses)); err != nil {
				return err
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
		if err := psf(monitor.WithField("secret", "sshkey")); err != nil {
			return err
		}

		pools, err := machinesSvc.ListPools()
		if err != nil {
			return err
		}

		for _, pool := range pools {
			if err := desireLb(pool); err != nil {
				return err
			}
			if err != nil {
				return err
			}
			machines, err := machinesSvc.List(pool)
			if err != nil {
				return err
			}
			for _, machine := range machines {
				if err := configureGcloud(machine, desired.JSONKey.Value); err != nil {
					return err
				}
			}
		}
		return nil
	}, addPools(current, desired, machinesSvc)
}
