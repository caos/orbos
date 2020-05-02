package static

import (
	"github.com/caos/orbiter/internal/push"
	"github.com/caos/orbiter/internal/secret"
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	dynamiclbmodel "github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/dynamic/wrap"

	//	externallbmodel "github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/external"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/ssh"
	"github.com/caos/orbiter/mntr"
)

func query(
	desired *DesiredV0,
	current *Current,

	nodeAgentsDesired map[string]*common.NodeAgentSpec,
	lb interface{},
	masterkey string,

	monitor mntr.Monitor,
	id string,
) (ensureFunc orbiter.EnsureFunc, err error) {

	// TODO: Allow Changes
	desireHostnameFunc := desireHostname(desired.Spec.Pools, nodeAgentsDesired)

	machinesSvc := NewMachinesService(monitor, desired, []byte(desired.Spec.Keys.BootstrapKeyPrivate.Value), []byte(desired.Spec.Keys.MaintenanceKeyPrivate.Value), []byte(desired.Spec.Keys.MaintenanceKeyPublic.Value), id, desireHostnameFunc)
	pools, err := machinesSvc.ListPools()
	if err != nil {
		return nil, err
	}

	current.Current.Ingresses = make(map[string]infra.Address)
	var desireLb func(pool string) error
	switch lbCurrent := lb.(type) {
	case *dynamiclbmodel.Current:

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

	for _, pool := range pools {

		if err := desireLb(pool); err != nil {
			return nil, err
		}

		machines, err := machinesSvc.List(pool)
		if err != nil {
			return nil, err
		}
		for _, machine := range machines {
			if err := desireHostnameFunc(machine, pool); err != nil {
				return nil, err
			}
		}
	}

	return func(psf push.Func) error {
		if (desired.Spec.Keys.MaintenanceKeyPrivate == nil || desired.Spec.Keys.MaintenanceKeyPrivate.Value == "") &&
			(desired.Spec.Keys.MaintenanceKeyPublic == nil || desired.Spec.Keys.MaintenanceKeyPublic.Value == "") {
			priv, pub, err := ssh.Generate()
			if err != nil {
				return err
			}
			desired.Spec.Keys.MaintenanceKeyPrivate = &secret.Secret{Masterkey: masterkey, Value: priv}
			desired.Spec.Keys.MaintenanceKeyPublic = &secret.Secret{Masterkey: masterkey, Value: pub}
			if err := psf(monitor.WithField("type", "maintenancekey")); err != nil {
				return err
			}
		}
		return nil
	}, addPools(current, desired, machinesSvc)
}
