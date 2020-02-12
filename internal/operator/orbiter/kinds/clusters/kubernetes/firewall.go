package kubernetes

import (
	"fmt"

	"github.com/caos/orbiter/internal/operator/common"
)

func firewallFuncs(desired DesiredV0, kubeAPIPort uint16) (desire func(machine initializedMachine), ensure func(machines []initializedMachine) bool) {

	desireFirewall := func(machine initializedMachine) common.Firewall {

		fw := map[string]common.Allowed{
			"kubelet": common.Allowed{
				Port:     fmt.Sprintf("%d", 10250),
				Protocol: "tcp",
			},
		}

		if machine.tier == Workers {
			fw["node-ports"] = common.Allowed{
				Port:     fmt.Sprintf("%d-%d", 30000, 32767),
				Protocol: "tcp",
			}
		}

		if machine.tier == Controlplane {
			fw["kubeapi-external"] = common.Allowed{
				Port:     fmt.Sprintf("%d", kubeAPIPort),
				Protocol: "tcp",
			}
			fw["kubeapi-internal"] = common.Allowed{
				Port:     fmt.Sprintf("%d", 6666),
				Protocol: "tcp",
			}
			fw["etcd"] = common.Allowed{
				Port:     fmt.Sprintf("%d-%d", 2379, 2380),
				Protocol: "tcp",
			}
			fw["kube-scheduler"] = common.Allowed{
				Port:     fmt.Sprintf("%d", 10251),
				Protocol: "tcp",
			}
			fw["kube-controller"] = common.Allowed{
				Port:     fmt.Sprintf("%d", 10252),
				Protocol: "tcp",
			}
		}

		if desired.Spec.Networking.Network == "calico" {
			fw["calico-bgp"] = common.Allowed{
				Port:     fmt.Sprintf("%d", 179),
				Protocol: "tcp",
			}
		}

		if machine.desiredNodeagent.Firewall == nil {
			machine.desiredNodeagent.Firewall = &common.Firewall{}
		}
		firewall := common.Firewall(fw)
		machine.desiredNodeagent.Firewall.Merge(firewall)
		return firewall
	}

	return func(machine initializedMachine) {
			desireFirewall(machine)
		}, func(machines []initializedMachine) bool {
			ready := true
			for _, machine := range machines {

				firewall := desireFirewall(machine)

				if machine.currentNodeagent == nil {
					ready = false
				} else if ready {
					ready = machine.currentNodeagent.Open.Contains(firewall)
				}
			}
			return ready
		}
}
