package main

import (
	"github.com/caos/orbos/internal/operator/boom/api"
	"github.com/caos/orbos/internal/secret"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"os"

	"github.com/spf13/cobra"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/orb"
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
					return api.SecretsFunc(orbconfig)
				} else if operator == "orbiter" {
					return orb.SecretsFunc(orbconfig)
				}
				return func(monitor mntr.Monitor, desiredTree *tree.Tree) (secrets map[string]*secret.Secret, err error) {
					return make(map[string]*secret.Secret), nil
				}
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
