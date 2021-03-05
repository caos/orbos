package main

import (
	"errors"
	"os"

	orbcfg "github.com/caos/orbos/pkg/orb"

	"github.com/caos/orbos/pkg/secret"

	"github.com/caos/orbos/internal/secret/operators"

	"github.com/spf13/cobra"
)

func ReadSecretCommand(getRv GetRootValues) *cobra.Command {

	return &cobra.Command{
		Use:     "readsecret [path]",
		Short:   "Print a secrets decrypted value to stdout",
		Long:    "Print a secrets decrypted value to stdout.\nIf no path is provided, a secret can interactively be chosen from a list of all possible secrets",
		Args:    cobra.MaximumNArgs(1),
		Example: `orbctl readsecret orbiter.k8s.kubeconfig > ~/.kube/config`,
		RunE: func(cmd *cobra.Command, args []string) (err error) {

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

			if !rv.Gitops {
				return errors.New("readsecret command is only supported with the --gitops flag yet")
			}

			if err := orbcfg.IsComplete(orbConfig); err != nil {
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
				path,
				operators.GetAllSecretsFunc(orbConfig))
			if err != nil {
				return err
			}
			if _, err := os.Stdout.Write([]byte(value)); err != nil {
				panic(err)
			}
			return nil
		},
	}
}
