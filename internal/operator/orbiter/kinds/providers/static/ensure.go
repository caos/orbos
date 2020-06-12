package static

import (
	"github.com/caos/orbos/internal/push"
	"github.com/caos/orbos/internal/secret"
	"github.com/pkg/errors"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	dynamiclbmodel "github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic/wrap"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/ssh"
	"github.com/caos/orbos/mntr"
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

	current.Current.Ingresses = make(map[string]*infra.Address)
	var desireLb func(pool string) error
	switch lbCurrent := lb.(type) {
	case *dynamiclbmodel.Current:

		mapVIP := func(vip *dynamiclbmodel.VIP) string {
			return vip.IP
		}

		desireLb = func(pool string) error {
			return lbCurrent.Current.Desire(pool, machinesSvc, nodeAgentsDesired, true, nil, mapVIP)
		}

		for _, pool := range lbCurrent.Current.Spec {
			for _, vip := range pool {
				for _, src := range vip.Transport {
					current.Current.Ingresses[src.Name] = &infra.Address{
						Location:     vip.IP,
						FrontendPort: uint16(src.FrontendPort),
						BackendPort:  uint16(src.BackendPort),
					}
				}
			}
		}

		machinesSvc = wrap.MachinesService(machinesSvc, *lbCurrent, nodeAgentsDesired, true, nil, mapVIP)
		//	case *externallbmodel.Current:
		//		for name, address := range lbCurrent.Current.Addresses {
		//			current.Current.Ingresses[name] = address
		//		}
	default:
		return nil, errors.Errorf("Unknown load balancer of type %T", lb)
	}

	for _, pool := range pools {
		desireLbFunc := func() error {
			return desireLb(pool)
		}
		if err := orbiter.EnsureFuncGoroutine(desireLbFunc); err != nil {
			return nil, err
		}

		machines, err := machinesSvc.List(pool)
		if err != nil {
			return nil, err
		}
		for _, machine := range machines {
			desireHostnameFuncFunc := func() error {
				return desireHostnameFunc(machine, pool)
			}
			if err := orbiter.EnsureFuncGoroutine(desireHostnameFuncFunc); err != nil {
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
