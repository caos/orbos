package cs

import (
	"fmt"

	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	dynamiclbmodel "github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/loadbalancers/dynamic/wrap"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/providers/core"
	"github.com/caos/orbos/mntr"
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
		panic(fmt.Errorf("unknown or unsupported load balancing of type %T", lb))
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
		_, err := core.DesireInternalOSFirewall(context.monitor, nodeAgentsDesired, nodeAgentsCurrent, context.machinesService, true, []string{"eth0"})
		if err != nil {
			return err
		}

		return ensureServer(context, current, hostPools, pool, m.(*machine), ensureNodeAgent)
	}
	wrappedMachines := wrap.MachinesService(context.machinesService, *lbCurrent, &dynamiclbmodel.VRRP{
		VRRPInterface: "eth1",
		VIPInterface:  "eth0",
		NotifyMaster:  notifyMaster(hostPools, current, poolsWithUnassignedVIPs),
		AuthCheck:     checkAuth,
	}, desiredToCurrentVIP(current))
	return func(pdf func(mntr.Monitor) error) *orbiter.EnsureResult {
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

				fwDone, err := core.DesireInternalOSFirewall(context.monitor, nodeAgentsDesired, nodeAgentsCurrent, context.machinesService, true, []string{"eth0"})
				if err != nil {
					return err
				}
				/* TODO: Remove unused code
				vips, err := allHostedVIPs(hostPools, context.machinesService, current)
				if err != nil {
					return err
				}
				nwDone, err := core.DesireOSNetworking(context.monitor, nodeAgentsDesired, nodeAgentsCurrent, context.machinesService, "dummy", vips)
				if err != nil {
					return err
				}
				*/
				done = lbDone && fwDone //&& nwDone
				return nil
			},
		})())
	}, addPools(current, desired, wrappedMachines)
}
