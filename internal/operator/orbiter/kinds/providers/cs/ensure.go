package cs

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	dynamiclbmodel "github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic/wrap"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"
)

func query(
	desired *Spec,
	current *Current,
	lb interface{},
	context *context,
	nodeAgentsCurrent *common.CurrentNodeAgents,
	nodeAgentsDesired *common.DesiredNodeAgents,
	naFuncs core.IterateNodeAgentFuncs,
	orbiterCommit string,
) (ensureFunc orbiter.EnsureFunc, err error) {

	lbCurrent, ok := lb.(*dynamiclbmodel.Current)
	if !ok {
		panic(errors.Errorf("Unknown or unsupported load balancing of type %T", lb))
	}

	hostPools, err := lbCurrent.Current.Spec(context.machinesService)
	if err != nil {
		return nil, err
	}

	ensureFIPs, removeFIPs, err := queryFloatingIPs(context, hostPools, current)
	if err != nil {
		return nil, err
	}

	queryNA, installNA := naFuncs(nodeAgentsCurrent)
	ensureNodeAgent := func(m infra.Machine) error {
		running, err := queryNA(m, orbiterCommit)
		if err != nil {
			return err
		}
		if !running {
			return installNA(m)
		}
		return nil
	}

	ensureServers, err := queryServers(context, hostPools, ensureNodeAgent)
	if err != nil {
		return nil, err
	}

	context.machinesService.onCreate = func(pool string, m infra.Machine) error {
		return ensureServer(context, hostPools, pool, m.(*machine), ensureNodeAgent)
	}
	wrappedMachines := wrap.MachinesService(context.machinesService, *lbCurrent, true, func(machine infra.Machine, peers infra.Machines, vips []*dynamiclbmodel.VIP) string {
		return ""
	}, func(vip *dynamic.VIP) string {
		for idx := range vip.Transport {
			transport := vip.Transport[idx]
			address, ok := current.Current.Ingresses[transport.Name]
			if ok {
				return address.Location
			}
		}
		panic(fmt.Errorf("external address for %v is not ensured", vip))
	})
	return func(pdf api.PushDesiredFunc) *orbiter.EnsureResult {
		var done bool
		return orbiter.ToEnsureResult(done, helpers.Fanout([]func() error{
			func() error { return helpers.Fanout(ensureFIPs)() },
			func() error { return helpers.Fanout(removeFIPs)() },
			func() error { return helpers.Fanout(ensureServers)() },
			func() error {
				var err error
				done, err = wrappedMachines.InitializeDesiredNodeAgents()
				return err
			},
		})())
	}, addPools(current, desired, wrappedMachines)
}
