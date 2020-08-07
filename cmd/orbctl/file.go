package main

import (
	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/internal/orb"
	"github.com/spf13/cobra"
)

func FileCommand() *cobra.Command {

	return &cobra.Command{
		Use:     "file <path> [command]",
		Short:   "Work with an orbs remote repository file",
		Example: `orbctl file <print|patch|edit|overwrite> orbiter.yml `,
	}
}

func initRepo(orbConfig *orb.Orb, gitClient *git.Client) error {
	if err := orbConfig.IsConnectable(); err != nil {
		return err
	}

	if err := gitClient.Configure(orbConfig.URL, []byte(orbConfig.Repokey)); err != nil {
		return err
	}

	return gitClient.Clone()
}
