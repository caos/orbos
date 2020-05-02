package gce

import (
	"github.com/caos/orbiter/internal/push"
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/orbiter"

	//	externallbmodel "github.com/caos/orbiter/internal/operator/orbiter/kinds/loadbalancers/external"

	"github.com/caos/orbiter/mntr"
)

func query(
	desired *Desired,
	current *Current,

	nodeAgentsDesired map[string]*common.NodeAgentSpec,
	lb interface{},
	masterkey string,

	monitor mntr.Monitor,
	id string,
) (ensureFunc orbiter.EnsureFunc, err error) {
	return func(_ push.Func) error { return nil }, errors.New("Not yet implemented")
	/*
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

		return func(_ push.Func) error { return nil }, addPools(current, desired, machinesSvc)*/
}
