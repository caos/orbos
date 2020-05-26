package cmd

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/mntr"
)

func Reconcile(monitor mntr.Monitor, kubeconfig *string, version string) error {
	k8sClient := kubernetes.NewK8sClient(monitor, kubeconfig)
	if *kubeconfig == "" {
		err := k8sClient.RefreshLocal()
		if err != nil {
			return err
		}
	}

	if k8sClient.Available() {
		if err := kubernetes.EnsureBoomArtifacts(monitor, k8sClient, version); err != nil {
			monitor.Info("Failed to deploy boom into k8s-cluster")
			return err
		}
		monitor.Info("Applied boom")
	} else {
		monitor.Info("Failed to connect to k8s")
	}

	return nil
}
