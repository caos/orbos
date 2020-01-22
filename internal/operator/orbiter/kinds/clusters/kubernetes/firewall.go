package kubernetes

import (
	"fmt"

	"github.com/caos/orbiter/internal/operator/common"
)

func ensureFirewall(currentComputes map[string]*Compute, nodeAgentsDesired map[string]*common.NodeAgentSpec, nodeAgentsCurrent map[string]*common.NodeAgentCurrent, desired DesiredV0, kubeAPIPort uint16) (bool, error) {

	ready := true
	for id, current := range currentComputes {

		fw := map[string]common.Allowed{
			"kubelet": common.Allowed{
				Port:     fmt.Sprintf("%d", 10250),
				Protocol: "tcp",
			},
		}

		if current.Metadata.Tier == Workers {
			fw["node-ports"] = common.Allowed{
				Port:     fmt.Sprintf("%d-%d", 30000, 32767),
				Protocol: "tcp",
			}
		}

		if current.Metadata.Tier == Controlplane {
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

		firewall := common.Firewall(fw)
		nodeAgentsDesired[id].Firewall.Merge(firewall)

		curr, ok := nodeAgentsCurrent[id]
		if ready && ok {
			ready = curr.Open.Contains(firewall)
		}
	}

	return ready, nil
}
