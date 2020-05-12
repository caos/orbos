package gce

import (
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	dynamiclbmodel "github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/dynamic"
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

	lbCurrent, ok := lb.(*dynamiclbmodel.Current)
	if !ok {
		errors.Errorf("Unknown or unsupported load balancing of type %T", lb)
	}

	return func(psf push.Func) error {
		externalIPs, err := addressesSvc.ensure(lbCurrent.Current.Spec)
		if err != nil {
			return err
		}

		current.Current.Ingresses = ingresses

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
