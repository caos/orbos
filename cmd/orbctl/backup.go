package main

import (
	"github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/start"
	"github.com/spf13/cobra"
)

func BackupCommand(rv RootValues) *cobra.Command {
	var (
		backup string
		cmd    = &cobra.Command{
			Use:   "backup",
			Short: "Instant backup",
			Long:  "Instant backup",
		}
	)

	flags := cmd.Flags()
	flags.StringVar(&backup, "backup", "", "Name used for backup folder")

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {
		_, monitor, orbConfig, gitClient, errFunc, err := rv()
		if err != nil {
			return err
		}
		defer func() {
			err = errFunc(err)
		}()

		if err := orbConfig.IsConnectable(); err != nil {
			return err
		}

		if err := gitClient.Configure(orbConfig.URL, []byte(orbConfig.Repokey)); err != nil {
			return err
		}

		if err := gitClient.Clone(); err != nil {
			return err
		}

		found, err := api.ExistsZitadelYml(gitClient)
		if err != nil {
			return err
		}
		if found {
			kubeconfigs, err := start.GetKubeconfigs(monitor, gitClient, orbConfig)
			if err != nil {
				return err
			}
			for _, kubeconfig := range kubeconfigs {
				k8sClient := kubernetes.NewK8sClient(monitor, &kubeconfig)
				if k8sClient.Available() {
					err := start.ZitadelBackup(monitor, orbConfig.Path, k8sClient, backup)
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
