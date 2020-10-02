package cmds

import (
	"github.com/caos/orbos/internal/api"
	boomapi "github.com/caos/orbos/internal/operator/boom/api"
	cmdboom "github.com/caos/orbos/internal/operator/boom/cmd"
	orbdb "github.com/caos/orbos/internal/operator/database/kinds/orb"
	orbnw "github.com/caos/orbos/internal/operator/networking/kinds/orb"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/git"
	kubernetes2 "github.com/caos/orbos/pkg/kubernetes"
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

	k8sClient := kubernetes2.NewK8sClient(monitor, kubeconfig)

	if err := cmdboom.Reconcile(monitor, k8sClient, version, desiredKind.Spec.Boom); err != nil {
		return err
	}
	return nil
}

func deployDatabase(monitor mntr.Monitor, gitClient *git.Client, kubeconfig *string) error {
	found, err := api.ExistsDatabaseYml(gitClient)
	if err != nil {
		return err
	}
	if found {
		k8sClient := kubernetes2.NewK8sClient(monitor, kubeconfig)

		if k8sClient.Available() {
			tree, err := api.ReadDatabaseYml(gitClient)
			if err != nil {
				return err
			}

			if err := orbdb.Reconcile(monitor, tree)(k8sClient); err != nil {
				return err
			}
		} else {
			monitor.Info("Failed to connect to k8s")
		}
	}
	return nil
}

func deployNetworking(monitor mntr.Monitor, gitClient *git.Client, kubeconfig *string) error {
	found, err := api.ExistsNetworkingYml(gitClient)
	if err != nil {
		return err
	}
	if found {
		k8sClient := kubernetes2.NewK8sClient(monitor, kubeconfig)

		if k8sClient.Available() {
			tree, err := api.ReadNetworkinglYml(gitClient)
			if err != nil {
				return err
			}

			if err := orbnw.Reconcile(monitor, tree)(k8sClient); err != nil {
				return err
			}
		} else {
			monitor.Info("Failed to connect to k8s")
		}
	}
	return nil
}
