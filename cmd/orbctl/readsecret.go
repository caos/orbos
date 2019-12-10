package main

import (
	"os"

	"github.com/caos/orbiter/internal/core/secret"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func readSecretCommand(rv rootValues) *cobra.Command {

	return &cobra.Command{
		Use:   "readsecret [name]",
		Short: "Decrypt and print to stdout",
		Args:  cobra.ExactArgs(1),
		Example: `
mkdir -p ~/.kube
orbctl --repourl git@github.com:example/my-orb.git \
       --repokey-file ~/.ssh/my-orb --masterkey 'my very secret key'
	   readsecret myorbk8s_kubeconfig > ~/.kube/config`,
		RunE: func(cmd *cobra.Command, args []string) error {

			_, logger, gitClient, orb, errFunc := rv()
			if errFunc != nil {
				return errFunc(cmd)
			}

			if err := gitClient.Clone(); err != nil {
				panic(err)
			}

			sec, err := gitClient.Read("secrets.yml")
			if err != nil {
				panic(err)
			}

			var secsMap map[string]interface{}
			if err := yaml.Unmarshal(sec, &secsMap); err != nil {
				panic(errors.Wrap(err, "Unmarshalling failed"))
			}

			if err := secret.New(logger, secsMap, args[0], orb.Masterkey).Read(os.Stdout); err != nil {
				panic(err)
			}
			return nil
		},
	}
}
