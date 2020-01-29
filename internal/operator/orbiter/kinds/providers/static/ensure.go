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
	sec *SecretsV0,

	psf orbiter.PushSecretsFunc,
	nodeAgentsDesired map[string]*common.NodeAgentSpec,
	lb interface{},
	masterkey string,

	logger logging.Logger,
	id string,
) (err error) {

	if (sec.Secrets.MaintenanceKeyPrivate == nil || sec.Secrets.MaintenanceKeyPrivate.Value == "") &&
		(sec.Secrets.MaintenanceKeyPublic == nil || sec.Secrets.MaintenanceKeyPublic.Value == "") {
		priv, pub, err := ssh.Generate()
		if err != nil {
			return err
		}
		sec.Secrets.MaintenanceKeyPrivate = &orbiter.Secret{Masterkey: masterkey, Value: priv}
		sec.Secrets.MaintenanceKeyPublic = &orbiter.Secret{Masterkey: masterkey, Value: pub}
		if err := psf(); err != nil {
			return err
		}
	}

	// TODO: Allow Changes
	desireHostnameFunc := desireHostname(desired.Spec.Pools, nodeAgentsDesired)

	computesSvc := NewComputesService(logger, desired, []byte(sec.Secrets.BootstrapKeyPrivate.Value), []byte(sec.Secrets.MaintenanceKeyPrivate.Value), []byte(sec.Secrets.MaintenanceKeyPublic.Value), id, desireHostnameFunc)
	pools, err := computesSvc.ListPools()
	if err != nil {
		return err
	}
	for _, pool := range pools {
		computes, err := computesSvc.List(pool, true)
		if err != nil {
			return err
		}
		for _, compute := range computes {
			if err := desireHostnameFunc(compute, pool); err != nil {
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
			if err := lbCurrent.Current.Desire(pool, computesSvc, nodeAgentsDesired, ""); err != nil {
				return err
			}
		}
		computesSvc = wrap.ComputesService(computesSvc, *lbCurrent, nodeAgentsDesired, "")
		//	case *externallbmodel.Current:
		//		for name, address := range lbCurrent.Current.Addresses {
		//			current.Current.Ingresses[name] = address
		//		}
	default:
		return errors.Errorf("Unknown load balancer of type %T", lb)
	}

	return addPools(current, desired, computesSvc)
}
