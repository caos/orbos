package main

import (
	"fmt"
	"github.com/caos/orbos/internal/api"
	"github.com/caos/orbos/internal/operator/database/kinds/orb"
	"github.com/spf13/cobra"
)

func BackupListCommand(getRv GetRootValues) *cobra.Command {
	var (
		cmd = &cobra.Command{
			Use:   "backuplist",
			Short: "Get a list of all backups",
			Long:  "Get a list of all backups",
		}
	)

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {

		rv, err := getRv()
		if err != nil {
			return err
		}
		defer func() {
			err = rv.ErrFunc(err)
		}()

		monitor := rv.Monitor
		orbConfig := rv.OrbConfig
		gitClient := rv.GitClient

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
			zitadel, err := api.ReadDatabaseYml(gitClient)
			if err != nil {
				return err
			}
			list, err := orb.BackupListFunc()(monitor, zitadel)
			if err != nil {
				return err
			}
			for _, v := range list {
				fmt.Println(v)
			}
		}
		return nil
	}
	return cmd
}
