package main

import (
	"github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/internal/start"
	kubernetes2 "github.com/caos/orbos/pkg/kubernetes"
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

		if err := gitClient.Clone(); err != nil {
			return err
		}

		found, err := api.ExistsDatabaseYml(gitClient)
		if err != nil {
			return err
		}
		if found {
			kubeconfigs, err := start.GetKubeconfigs(monitor, gitClient, orbConfig, version)
			if err != nil {
				return err
			}
			for _, kubeconfig := range kubeconfigs {
				k8sClient := kubernetes2.NewK8sClient(monitor, &kubeconfig)
				if k8sClient.Available() {
					if err := start.DatabaseBackup(monitor, orbConfig.Path, k8sClient, backup, &version); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	return cmd
}
