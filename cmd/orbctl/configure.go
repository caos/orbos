package main

import (
	"errors"
	"fmt"

	"github.com/caos/orbos/v5/mntr"

	"github.com/caos/orbos/v5/pkg/cfg"

	"github.com/caos/orbos/v5/pkg/kubernetes/cli"

	"github.com/caos/orbos/v5/pkg/git"

	"github.com/caos/orbos/v5/pkg/orb"
	"github.com/spf13/cobra"
)

func ConfigCommand(getRv GetRootValues) *cobra.Command {

	var (
		newMasterKey string
		newRepoURL   string
		newRepoKey   string
		cmd          = &cobra.Command{
			Use:     "configure",
			Short:   "Configures and reconfigures an orb",
			Long:    "Generates missing ssh keys and other secrets where it makes sense",
			Aliases: []string{"reconfigure", "config", "reconfig"},
		}
	)

	flags := cmd.Flags()
	flags.StringVar(&newMasterKey, "masterkey", "", "Reencrypts all secrets")
	flags.StringVar(&newRepoURL, "repourl", "", "Configures the repository URL")
	flags.StringVar(&newRepoKey, "repokey", "", "Configures the used key to communicate with the repository")

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {

		rv, _ := getRv("configure", "", map[string]interface{}{"masterkey": newMasterKey != "", "newRepoURL": newRepoURL})
		defer rv.ErrFunc(err)

		if !rv.Gitops {
			return mntr.ToUserError(errors.New("configure command is only supported with the --gitops flag"))
		}

		if err := orb.Reconfigure(
			rv.Ctx,
			rv.Monitor,
			rv.OrbConfig,
			newRepoURL,
			newMasterKey,
			newRepoKey,
			rv.GitClient,
			githubClientID,
			githubClientSecret,
		); err != nil {
			return err
		}

		k8sClient, err := cli.Client(rv.Monitor, rv.OrbConfig, rv.GitClient, rv.Kubeconfig, rv.Gitops, false)
		if err != nil {
			// ignore
			err = nil
		}

		unmanagedOperators := []git.DesiredFile{git.DatabaseFile, git.ZitadelFile}
		for i := range unmanagedOperators {
			operatorFile := unmanagedOperators[i]
			if rv.GitClient.Exists(operatorFile) {
				return mntr.ToUserError(fmt.Errorf("found %s in git repository. Please use zitadelctl's configure command", operatorFile))
			}
		}

		if err := cfg.ApplyOrbconfigSecret(
			rv.OrbConfig,
			k8sClient,
			rv.Monitor,
		); err != nil {
			return err
		}

		return cfg.ConfigureOperators(
			rv.GitClient,
			rv.OrbConfig.Masterkey,
			cfg.ORBOSConfigurers(
				rv.Ctx,
				rv.Monitor,
				rv.OrbConfig,
				rv.GitClient,
			))
	}
	return cmd
}
