package kubernetes

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbos/mntr"
)

func destroy(monitor mntr.Monitor, providerCurrents map[string]interface{}, k8sClient *Client) error {

	if k8sClient.Available() {
		k8sClient.DeleteDeployment("caos-system", "orbiter")
	}

	for _, provider := range providerCurrents {
		prov := provider.(infra.ProviderCurrent)
		for _, pool := range prov.Pools() {
			machines, err := pool.GetMachines(false)
			if err != nil {
				return err
			}
			for _, machine := range machines {
				machine.Execute(nil, nil, "sudo systemctl stop node-agentd")
				machine.Execute(nil, nil, "sudo systemctl disable node-agentd")
				machine.Execute(nil, nil, "sudo kubeadm reset -f")
				machine.Execute(nil, nil, "sudo rm -rf /var/lib/etcd")
			}
		}
	}
	return nil
}
