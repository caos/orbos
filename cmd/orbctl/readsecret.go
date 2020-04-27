package main

import (
	"github.com/caos/orbiter/internal/operator/boom/api"
	"github.com/caos/orbiter/internal/secret"
	"os"

	"github.com/spf13/cobra"

	"github.com/caos/orbiter/internal/operator/orbiter/kinds/orb"
)

func ReadSecretCommand(rv RootValues) *cobra.Command {

	return &cobra.Command{
		Use:     "readsecret [path]",
		Short:   "Print a secrets decrypted value to stdout",
		Long:    "Print a secrets decrypted value to stdout.\nIf no path is provided, a secret can interactively be chosen from a list of all possible secrets",
		Args:    cobra.MaximumNArgs(1),
		Example: `orbctl readsecret orbiter.k8s.kubeconfig > ~/.kube/config`,
		RunE: func(cmd *cobra.Command, args []string) error {

			_, logger, gitClient, orbconfig, errFunc := rv()
			if errFunc != nil {
				return errFunc(cmd)
			}

			path := ""
			if len(args) > 0 {
				path = args[0]
			}

			secretFunc := func(operator string) secret.Func {
				if operator == "boom" {
					return api.SecretFunc(orbconfig)
				} else if operator == "orbiter" {
					return orb.SecretsFunc(orbconfig)
				}
				return nil
			}

			value, err := secret.Read(
				logger,
				gitClient,
				secretFunc,
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
