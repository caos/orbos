package static

import (
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	dynamiclbmodel "github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/dynamic/wrap"

	//	externallbmodel "github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/external"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/providers/static/ssh"
	"github.com/caos/orbiter/logging"
)

func ensure(
	desired *DesiredV0,
	current *Current,

	psf orbiter.PushSecretsFunc,
	nodeAgentsDesired map[string]*common.NodeAgentSpec,
	lb interface{},
	masterkey string,

	logger logging.Logger,
	id string,
) (err error) {

	if (desired.Spec.Keys.MaintenanceKeyPrivate == nil || desired.Spec.Keys.MaintenanceKeyPrivate.Value == "") &&
		(desired.Spec.Keys.MaintenanceKeyPublic == nil || desired.Spec.Keys.MaintenanceKeyPublic.Value == "") {
		priv, pub, err := ssh.Generate()
		if err != nil {
			return err
		}
		desired.Spec.Keys.MaintenanceKeyPrivate = &orbiter.Secret{Masterkey: masterkey, Value: priv}
		desired.Spec.Keys.MaintenanceKeyPublic = &orbiter.Secret{Masterkey: masterkey, Value: pub}
		if err := psf(); err != nil {
			return err
		}
	}

	// TODO: Allow Changes
	desireHostnameFunc := desireHostname(desired.Spec.Pools, nodeAgentsDesired)

	machinesSvc := NewMachinesService(logger, desired, []byte(desired.Spec.Keys.BootstrapKeyPrivate.Value), []byte(desired.Spec.Keys.MaintenanceKeyPrivate.Value), []byte(desired.Spec.Keys.MaintenanceKeyPublic.Value), id, desireHostnameFunc)
	pools, err := machinesSvc.ListPools()
	if err != nil {
		return err
	}
	for _, pool := range pools {
		machines, err := machinesSvc.List(pool, true)
		if err != nil {
			return err
		}
		for _, machine := range machines {
			if err := desireHostnameFunc(machine, pool); err != nil {
				return err
			}
		}
	}

	current.Current.Ingresses = make(map[string]infra.Address)
	switch lbCurrent := lb.(type) {
	case *dynamiclbmodel.Current:
		for name, address := range lbCurrent.Current.Addresses {
			current.Current.Ingresses[name] = address
		}
		for _, pool := range pools {
			if err := lbCurrent.Current.Desire(pool, machinesSvc, nodeAgentsDesired, ""); err != nil {
				return err
			}
		}
		machinesSvc = wrap.MachinesService(machinesSvc, *lbCurrent, nodeAgentsDesired, "")
		//	case *externallbmodel.Current:
		//		for name, address := range lbCurrent.Current.Addresses {
		//			current.Current.Ingresses[name] = address
		//		}
	default:
		return errors.Errorf("Unknown load balancer of type %T", lb)
	}

	return addPools(current, desired, machinesSvc)
}
