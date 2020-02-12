package kubernetes

import (
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/kubernetes/edge/k8s"
	"github.com/caos/orbiter/logging"
)

func destroy(logger logging.Logger, providerCurrents map[string]interface{}, k8sClient *k8s.Client) error {

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
