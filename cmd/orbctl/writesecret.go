package main

import (
	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/core/secret"
	"github.com/caos/orbiter/internal/edge/git"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

func writeSecretCommand(rv rootValues) *cobra.Command {

	var (
		value string
		file  string
		stdin bool
		cmd   = &cobra.Command{
			Use:   "writesecret [name]",
			Short: "Encrypt and push",
			Args:  cobra.ExactArgs(1),
			Example: `alias myorb="orbctl --repourl git@github.com:example/my-orb.git --repokey-file ~/.ssh/my-orb-repo --masterkey 'my very secret key'"
myorb writesecret myorbsomeclusterstaticprovider_bootstrapkey --file ~/.ssh/my-orb-bootstrap
myorb writesecret myorbsomeclusterstaticprovider_bootstrapkey_pub --file ~/.ssh/my-orb-bootstrap.pub
myorb writesecret myorbsomeclustergceprovider_google_application_credentials_value --value "$(cat $GOOGLE_APPLICATION_CREDENTIALS)" `,
		}
	)

	flags := cmd.Flags()
	flags.StringVar(&value, "value", "", "Secret phrase value used for encrypting and decrypting secrets")
	flags.StringVarP(&file, "file", "f", "", "Secret phrase file used for encrypting and decrypting secrets")
	flags.BoolVar(&stdin, "stdin", false, "Read Secret phrase used for encrypting and decrypting secrets from standard input")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {

		rootFlags := cmd.InheritedFlags()
		mustGetBool := func(flag string) bool {
			v, err := rootFlags.GetBool(flag)
			if err != nil {
				panic(err)
			}
			return v
		}

		if stdin && (mustGetBool("masterkey-stdin") || mustGetBool("repokey-stdin")) {
			return errMultipleStdinKeys
		}

		s, err := key(value, file, stdin)
		if err != nil {
			return err
		}

		_, logger, gitClient, _, _, mk, err := rv()
		if err != nil {
			return err
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
			panic(err)
		}

		if err := secret.New(logger, secsMap, args[0], mk).Write([]byte(s)); err != nil {
			panic(err)
		}

		if _, err := gitClient.UpdateRemoteUntilItWorks(&git.File{
			Path: "secrets.yml",
			Overwrite: func(o []byte) ([]byte, error) {
				var newSecsMap map[string]interface{}
				if err := yaml.Unmarshal(o, &newSecsMap); err != nil {
					panic(err)
				}
				newSecsMap[args[0]] = secsMap[args[0]]
				return operator.Marshal(newSecsMap)
			},
			Force: true,
		}); err != nil {
			panic(err)
		}
		return nil
	}
	return cmd
}
