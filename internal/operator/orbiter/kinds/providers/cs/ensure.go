package cs

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
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

	ensureServers, err := queryServers(context, current, hostPools, ensureNodeAgent)
	if err != nil {
		return nil, err
	}

	context.machinesService.onCreate = func(pool string, m infra.Machine) error {
		return ensureServer(context, current, hostPools, pool, m.(*machine), ensureNodeAgent)
	}
	wrappedMachines := wrap.MachinesService(context.machinesService, *lbCurrent, "eth1", notifyMaster(hostPools, current, context), desiredToCurrentVIP(current))
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

func addDummyIPCommand(ips []string) string {

	if len(ips) == 0 {
		return "true"
	}

	cmd := "ip link add dummy1 type dummy"
	for idx := range ips {
		ip := ips[idx]
		if ip == "" {
			return "true"
		}
		cmd += fmt.Sprintf(" && ip addr add %s/32 dev dummy1", ip)
	}

	return cmd
}
