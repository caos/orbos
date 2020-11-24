package core

import (
	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/mntr"
)

func DesireInternalOSFirewall(monitor mntr.Monitor, nodeAgentsDesired *common.DesiredNodeAgents, nodeAgentsCurrent *common.CurrentNodeAgents, service MachinesService, openInterfaces []string) (bool, error) {
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
				"external": {Interfaces: openInterfaces},
			}})
	}
	return done, nil
}
