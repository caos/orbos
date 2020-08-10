package main

import (
	"github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/orb"
	"github.com/caos/orbos/internal/start"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
)

func RestoreCommand(rv RootValues) *cobra.Command {
	var (
		backup string
		cmd    = &cobra.Command{
			Use:   "restore",
			Short: "Restore from backup",
			Long:  "Restore from backup",
		}
	)

	flags := cmd.Flags()
	flags.StringVar(&backup, "backup", "", "Backup used for db restore")

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
			zitadel, err := api.ReadZitadelYml(gitClient)
			if err != nil {
				return err
			}
			list, err := orb.BackupListFunc()(monitor, zitadel)
			if err != nil {
				return err
			}

			if backup == "" {
				prompt := promptui.Select{
					Label: "Select machine",
					Items: list,
				}

				_, result, err := prompt.Run()
				if err != nil {
					return err
				}
				backup = result
			}

			kubeconfigs, err := start.GetKubeconfigs(monitor, gitClient)
			if err != nil {
				return err
			}
			for _, kubeconfig := range kubeconfigs {
				k8sClient := kubernetes.NewK8sClient(monitor, &kubeconfig)
				if k8sClient.Available() {
					if err := start.ZitadelRestore(monitor, orbConfig.Path, k8sClient, backup); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}
	return cmd
}
