package kubernetes

import (
	"fmt"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/mntr"
)

func firewallFunc(monitor mntr.Monitor, desired DesiredV0) (desire func(machine *initializedMachine)) {

	return func(machine *initializedMachine) {

		monitor = monitor.WithField("machine", machine.infra.ID())

		fw := map[string]*common.Allowed{
			"kubelet": {
				Port:     fmt.Sprintf("%d", 10250),
				Protocol: "tcp",
			},
		}

		if machine.pool.tier == Controlplane {
			fw["etcd"] = &common.Allowed{
				Port:     fmt.Sprintf("%d-%d", 2379, 2381),
				Protocol: "tcp",
			}
			fw["kube-scheduler"] = &common.Allowed{
				Port:     fmt.Sprintf("%d", 10251),
				Protocol: "tcp",
			}
			fw["kube-controller"] = &common.Allowed{
				Port:     fmt.Sprintf("%d", 10252),
				Protocol: "tcp",
			}
		}

		if desired.Spec.Networking.Network == "calico" {
			fw["calico-bgp"] = &common.Allowed{
				Port:     fmt.Sprintf("%d", 179),
				Protocol: "tcp",
			}
		}

		firewall := common.ToFirewall("internal", fw)
		firewallSources := common.Firewall{
			Zones: map[string]*common.Zone{
				"internal": {Sources: []string{
					string(desired.Spec.Networking.PodCidr),
					string(desired.Spec.Networking.ServiceCidr),
				}},
			},
		}
		firewall.Merge(firewallSources)

		machine.desiredNodeagent.Firewall.Merge(firewall)
		if firewall.IsContainedIn(machine.currentNodeagent.Open) {
			machine.currentMachine.FirewallIsReady = true
			monitor.Debug("firewall is ready")
			return
		}

		machine.currentMachine.FirewallIsReady = false
		monitor.WithField("open", firewall.AllZones()).Info("firewall desired")
	}
}
