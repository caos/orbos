package main

import (
	"os"

	"github.com/caos/orbos/internal/operator/secretfuncs"
	"github.com/caos/orbos/internal/secret"

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

			_, monitor, orbConfig, gitClient, errFunc := rv()
			if errFunc != nil {
				return errFunc(cmd)
			}

			if err := orbConfig.IsComplete(); err != nil {
				return err
			}

			if err := gitClient.Configure(orbConfig.URL, []byte(orbConfig.Repokey)); err != nil {
				return err
			}

			if err := gitClient.Clone(); err != nil {
				return err
			}

			path := ""
			if len(args) > 0 {
				path = args[0]
			}

			value, err := secret.Read(
				monitor,
				gitClient,
				secretfuncs.GetSecrets(),
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
