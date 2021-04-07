package main

import (
	"errors"

	"github.com/caos/orbos/pkg/git"

	orbcfg "github.com/caos/orbos/pkg/orb"

	"github.com/caos/orbos/internal/api"
	boomapi "github.com/caos/orbos/internal/operator/boom/api"
	"github.com/caos/orbos/internal/operator/orbiter"
	orbadapter "github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
	"github.com/caos/orbos/pkg/labels"
	"github.com/spf13/cobra"
)

func APICommand(getRv GetRootValues) *cobra.Command {
	var (
		cmd = &cobra.Command{
			Use:   "api",
			Short: "Upgrade the yml-files to the newest version",
			Long:  "Upgrade the yml-files to the newest version",
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

		if !rv.Gitops {
			return errors.New("api command is only supported with the --gitops flag")
		}

		monitor := rv.Monitor
		orbConfig := rv.OrbConfig
		gitClient := rv.GitClient

		if err := orbcfg.IsComplete(orbConfig); err != nil {
			return err
		}

		if err := gitClient.Configure(orbConfig.URL, []byte(orbConfig.Repokey)); err != nil {
			return err
		}

		if err := gitClient.Clone(); err != nil {
			return err
		}

		foundOrbiter, err := gitClient.Exists(git.OrbiterFile)
		if err != nil {
			return err
		}

		var desireds []api.GitDesiredState
		if foundOrbiter {
			_, _, _, migrate, desired, _, _, err := orbiter.Adapt(gitClient, monitor, make(chan struct{}), orbadapter.AdaptFunc(
				labels.NoopOperator("ORBOS"),
				orbConfig,
				gitCommit,
				true,
				false,
				gitClient,
			))
			if err != nil {
				return err
			}

			if migrate {
				desireds = append(desireds, api.GitDesiredState{
					Desired: desired,
					Path:    git.OrbiterFile,
				})
			}

		}
		foundBoom, err := gitClient.Exists(git.BoomFile)
		if err != nil {
			return err
		}
		if foundBoom {

			desired, err := gitClient.ReadTree(git.BoomFile)
			if err != nil {
				return err
			}

			toolset, migrate, _, _, err := boomapi.ParseToolset(desired)
			if err != nil {
				return err
			}
			if migrate {
				desired.Parsed = toolset
				desireds = append(desireds, api.GitDesiredState{
					Desired: desired,
					Path:    git.BoomFile,
				})
			}
		}
		if len(desireds) > 0 {
			return api.PushGitDesiredStates(monitor, "migrate apis", gitClient, desireds)
		}
		return nil
	}
	return cmd
}
