package cmds

import (
	"github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/internal/git"
	boomapi "github.com/caos/orbos/internal/operator/boom/api"
	cmdboom "github.com/caos/orbos/internal/operator/boom/cmd"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	cmdzitadel "github.com/caos/orbos/internal/operator/zitadel/cmd"
	"github.com/caos/orbos/mntr"
)

func deployBoom(monitor mntr.Monitor, gitClient *git.Client, kubeconfig *string, version string) error {
	foundBoom, err := api.ExistsBoomYml(gitClient)
	if err != nil {
		return err
	}
	if !foundBoom {
		monitor.Info("No BOOM deployed as no boom.yml present")
		return nil
	}
	desiredTree, err := api.ReadBoomYml(gitClient)
	if err != nil {
		return err
	}

	desiredKind, _, err := boomapi.ParseToolset(desiredTree)
	if err != nil {
		return err
	}

	k8sClient := kubernetes.NewK8sClient(monitor, kubeconfig)

	if err := cmdboom.Reconcile(monitor, k8sClient, version, true, desiredKind.Spec.Boom); err != nil {
		return err
	}
	return nil
}

func deployZitadel(monitor mntr.Monitor, gitClient *git.Client, kubeconfig *string, version string) error {
	found, err := api.ExistsZitadelYml(gitClient)
	if err != nil {
		return err
	}
	if found {
		k8sClient := kubernetes.NewK8sClient(monitor, kubeconfig)

		if k8sClient.Available() {
			tree, err := api.ReadZitadelYml(gitClient)
			if err != nil {
				return err
			}

			if err := cmdzitadel.Reconcile(monitor, tree, version)(k8sClient); err != nil {
				return err
			}
		} else {
			monitor.Info("Failed to connect to k8s")
		}
	}
	return nil
}
