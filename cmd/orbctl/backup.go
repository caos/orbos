package main

import (
	"github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/start"
	"github.com/spf13/cobra"
)

func BackupCommand(rv RootValues) *cobra.Command {
	var (
		cmd = &cobra.Command{
			Use:   "backup",
			Short: "Instant backup",
			Long:  "Instant backup",
		}
	)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		_, monitor, orbConfig, gitClient, errFunc := rv()
		if errFunc != nil {
			return errFunc(cmd)
		}

		if err := orbConfig.IsConnectable(); err != nil {
			return err
		}

		if err := gitClient.Configure(orbConfig.URL, []byte(orbConfig.Repokey)); err != nil {
			return err
		}

		found, err := api.ExistsZitadelYml(gitClient)
		if err != nil {
			return err
		}
		if found {
			kubeconfigs, err := start.GetKubeconfigs(monitor, gitClient)
			if err != nil {
				return err
			}
			for _, kubeconfig := range kubeconfigs {
				k8sClient := kubernetes.NewK8sClient(monitor, &kubeconfig)
				if k8sClient.Available() {
					err := start.ZitadelBackup(monitor, orbConfig.Path, k8sClient)
					if err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	return cmd
}
