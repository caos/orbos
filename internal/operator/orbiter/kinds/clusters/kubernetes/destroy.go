package kubernetes

import (
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
)

func destroy(providerCurrents map[string]interface{}, kubeconfig *orbiter.Secret) error {
	for _, provider := range providerCurrents {
		prov := provider.(infra.ProviderCurrent)
		for _, pool := range prov.Pools() {
			computes, err := pool.GetComputes(false)
			if err != nil {
				return err
			}
			for _, compute := range computes {
				if _, err := compute.Execute(nil, nil, "sudo kubeadm reset -f"); err != nil {
					return err
				}
				if _, err := compute.Execute(nil, nil, "sudo rm -rf /var/lib/etcd"); err != nil {
					return err
				}
			}
		}
	}
	kubeconfig.Value = ""
	return nil
}
