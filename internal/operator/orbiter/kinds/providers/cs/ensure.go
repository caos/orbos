package cs

import (
	"github.com/caos/orbos/mntr"
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

	hostPools, authChecks, err := lbCurrent.Current.Spec(context.machinesService)
	if err != nil {
		return nil, err
	}

	ensureFIPs, removeFIPs, poolsWithUnassignedVIPs, err := queryFloatingIPs(context, hostPools, current)
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

		_, err := desireNodeAgents(context.monitor, nodeAgentsDesired, nodeAgentsCurrent, context.machinesService)
		if err != nil {
			return err
		}

		return ensureServer(context, current, hostPools, pool, m.(*machine), ensureNodeAgent)
	}
	wrappedMachines := wrap.MachinesService(context.machinesService, *lbCurrent, &dynamiclbmodel.VRRP{
		VRRPInterface: "eth1",
		NotifyMaster:  notifyMaster(hostPools, current, poolsWithUnassignedVIPs),
		AuthCheck:     checkAuth,
	}, desiredToCurrentVIP(current))
	return func(pdf api.PushDesiredFunc) *orbiter.EnsureResult {
		var done bool
		return orbiter.ToEnsureResult(done, helpers.Fanout([]func() error{
			func() error {
				return helpers.Fanout(ensureTokens(context.monitor, []byte(desired.APIToken.Value), authChecks))()
			},
			func() error { return helpers.Fanout(ensureFIPs)() },
			func() error { return helpers.Fanout(removeFIPs)() },
			func() error { return helpers.Fanout(ensureServers)() },
			func() error {
				lbDone, err := wrappedMachines.InitializeDesiredNodeAgents()
				if err != nil {
					return err
				}

				csDone, err := desireNodeAgents(context.monitor, nodeAgentsDesired, nodeAgentsCurrent, context.machinesService)
				if err != nil {
					return err
				}
				done = lbDone && csDone
				return nil
			},
		})())
	}, addPools(current, desired, wrappedMachines)
}

func desireNodeAgents(monitor mntr.Monitor, nodeAgentsDesired *common.DesiredNodeAgents, nodeAgentsCurrent *common.CurrentNodeAgents, service core.MachinesService) (bool, error) {
	done := true
	desireNodeAgent := func(machine infra.Machine, fw common.Firewall) {
		machineMonitor := monitor.WithField("machine", machine.ID())
		deepNa, _ := nodeAgentsDesired.Get(machine.ID())
		deepNaCurr, _ := nodeAgentsCurrent.Get(machine.ID())

		deepNa.Firewall.Merge(fw)
		machineMonitor.WithField("ports", fw.AllZones()).Debug("Desired Cloudscale Firewall")
		if !fw.IsContainedIn(deepNaCurr.Open) {
			machineMonitor.WithField("ports", deepNa.Firewall.AllZones()).Info("Awaiting firewalld config")
			done = false
		}
	}

	pools, err := service.ListPools()
	if err != nil {
		return false, err
	}
	var machines infra.Machines
	var ips []string
	for _, pool := range pools {
		poolMachines, err := service.List(pool)
		if err != nil {
			return false, err
		}
		for _, machine := range poolMachines {
			machines = append(machines, machine)
			ips = append(ips, machine.IP()+"/32")
		}
	}
	for _, machine := range machines {
		desireNodeAgent(machine, common.Firewall{
			Zones: map[string]*common.Zone{
				"internal": {Sources: ips},
				"external": {Interfaces: []string{"eth0"}},
			}})
	}
	return done, nil
}
