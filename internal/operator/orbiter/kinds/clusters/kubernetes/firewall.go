package kubernetes

import (
	"fmt"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/mntr"
)

func firewallFunc(monitor mntr.Monitor, desired DesiredV0, kubeAPIPort uint16) (desire func(machine *initializedMachine)) {

	return func(machine *initializedMachine) {

		monitor = monitor.WithField("machine", machine.infra.ID())

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
		if firewall.IsContainedIn(machine.currentNodeagent.Open) && machine.desiredNodeagent.Firewall.Contains(firewall) {
			machine.currentMachine.FirewallIsReady = true
			monitor.Debug("Firewall is ready")
			return
		}

		machine.currentMachine.FirewallIsReady = false
		machine.desiredNodeagent.Firewall.Merge(firewall)
		monitor.Info("Firewall desired")
	}
}
