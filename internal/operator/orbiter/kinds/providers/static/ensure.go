package static

import (
	"sync"

	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic/wrap"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"
	"github.com/caos/orbos/internal/push"
	"github.com/caos/orbos/internal/secret"
	"github.com/pkg/errors"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	dynamiclbmodel "github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbos/internal/ssh"
	"github.com/caos/orbos/mntr"
)

func query(
	desired *DesiredV0,
	current *Current,

	nodeAgentsDesired *common.DesiredNodeAgents,
	nodeAgentsCurrent *common.CurrentNodeAgents,
	lb interface{},
	masterkey string,

	monitor mntr.Monitor,
	id,
	orbiterCommit,
	repoURL,
	repoKey string,
) (ensureFunc orbiter.EnsureFunc, err error) {

	// TODO: Allow Changes
	desireHostnameFunc := desireHostname(desired.Spec.Pools, nodeAgentsDesired, nodeAgentsCurrent, monitor)
	queryNA, installNA := core.NodeAgentFuncs(monitor, orbiterCommit, repoURL, repoKey, nodeAgentsCurrent)

	ensureNodeFunc := func(machine infra.Machine, pool string) error {
		running, err := queryNA(machine)
		if err != nil {
			return err
		}
		if !running {
			if err := installNA(machine); err != nil {
				return err
			}
		}
		_, err = desireHostnameFunc(machine, pool)
		return err
	}

	machinesSvc := NewMachinesService(monitor, desired, []byte(desired.Spec.Keys.BootstrapKeyPrivate.Value), []byte(desired.Spec.Keys.MaintenanceKeyPrivate.Value), []byte(desired.Spec.Keys.MaintenanceKeyPublic.Value), id, ensureNodeFunc)
	pools, err := machinesSvc.ListPools()
	if err != nil {
		return nil, err
	}

	current.Current.Ingresses = make(map[string]*infra.Address)
	ensureLBFunc := func() *orbiter.EnsureResult {
		return &orbiter.EnsureResult{
			Err:  nil,
			Done: true,
		}
	}
	switch lbCurrent := lb.(type) {
	case *dynamiclbmodel.Current:

		mapVIP := func(vip *dynamiclbmodel.VIP) string {
			return vip.IP
		}

		wrappedMachinesService := wrap.MachinesService(machinesSvc, *lbCurrent, true, nil, mapVIP)
		machinesSvc = wrappedMachinesService
		ensureLBFunc = func() *orbiter.EnsureResult {
			return orbiter.ToEnsureResult(wrappedMachinesService.InitializeDesiredNodeAgents())
		}
		deployPools, err := lbCurrent.Current.Spec(machinesSvc)
		if err != nil {
			return nil, err
		}
		for _, pool := range deployPools {
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

		//	case *externallbmodel.Current:
		//		for name, address := range lbCurrent.Current.Addresses {
		//			current.Current.Ingresses[name] = address
		//		}
	default:
		return nil, errors.Errorf("Unknown load balancer of type %T", lb)
	}

	return func(psf push.Func) *orbiter.EnsureResult {
		var wg sync.WaitGroup
		for _, pool := range pools {
			machines, listErr := machinesSvc.List(pool)
			if listErr != nil {
				err = helpers.Concat(err, listErr)
			}
			for _, machine := range machines {
				wg.Add(1)
				go func(m infra.Machine, p string) {
					err = helpers.Concat(err, ensureNodeFunc(m, p))
					wg.Done()
				}(machine, pool)
			}
		}

		wg.Wait()
		if err != nil {
			return orbiter.ToEnsureResult(false, err)
		}

		if (desired.Spec.Keys.MaintenanceKeyPrivate == nil || desired.Spec.Keys.MaintenanceKeyPrivate.Value == "") &&
			(desired.Spec.Keys.MaintenanceKeyPublic == nil || desired.Spec.Keys.MaintenanceKeyPublic.Value == "") {
			priv, pub, err := ssh.Generate()
			if err != nil {
				return orbiter.ToEnsureResult(false, err)
			}
			desired.Spec.Keys.MaintenanceKeyPrivate = &secret.Secret{Masterkey: masterkey, Value: priv}
			desired.Spec.Keys.MaintenanceKeyPublic = &secret.Secret{Masterkey: masterkey, Value: pub}
			if err := psf(monitor.WithField("type", "maintenancekey")); err != nil {
				return orbiter.ToEnsureResult(false, err)
			}
		}

		return ensureLBFunc()
	}, addPools(current, desired, machinesSvc)
}
