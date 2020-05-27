package main

import (
	"github.com/caos/orbos/internal/operator/secretfuncs"
	orbconfig "github.com/caos/orbos/internal/orb"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/utils/orbgit"
	"os"

	"github.com/spf13/cobra"
)

func ReadSecretCommand(rv RootValues) *cobra.Command {

	return &cobra.Command{
		Use:     "readsecret [path]",
		Short:   "Print a secrets decrypted value to stdout",
		Long:    "Print a secrets decrypted value to stdout.\nIf no path is provided, a secret can interactively be chosen from a list of all possible secrets",
		Args:    cobra.MaximumNArgs(1),
		Example: `orbctl readsecret orbiter.k8s.kubeconfig > ~/.kube/config`,
		RunE: func(cmd *cobra.Command, args []string) error {

			ctx, monitor, orbConfigPath, errFunc := rv()
			if errFunc != nil {
				return errFunc(cmd)
			}

			path := ""
			if len(args) > 0 {
				path = args[0]
			}

			orbConfig, err := orbconfig.ParseOrbConfig(orbConfigPath)
			if err != nil {
				return err
			}

			gitClientConf := &orbgit.Config{
				Comitter:  "orbctl",
				Email:     "orbctl@caos.ch",
				OrbConfig: orbConfig,
				Action:    "readsecret",
			}

			gitClient, cleanUp, err := orbgit.NewGitClient(ctx, monitor, gitClientConf)
			defer cleanUp()
			if err != nil {
				return err
			}

			value, err := secret.Read(
				monitor,
				gitClient,
				secretfuncs.Get(orbConfig),
				path)
			if err != nil {
				panic(err)
			}
			if _, err := os.Stdout.Write([]byte(value)); err != nil {
				panic(err)
			}
			return nil
		},
	}
}
