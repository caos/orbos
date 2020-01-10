package main

import (
	"io/ioutil"
	"os"

	"github.com/caos/orbiter/internal/core/operator/orbiter"
	"github.com/caos/orbiter/internal/core/secret"
	"github.com/caos/orbiter/internal/edge/git"
	"github.com/pkg/errors"
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
			Example: `orbctl writesecret myorbsomeclusterstaticprovider_bootstrapkey --file ~/.ssh/my-orb-bootstrap
orbctl writesecret myorbsomeclusterstaticprovider_bootstrapkey_pub --file ~/.ssh/my-orb-bootstrap.pub
orbctl writesecret myorbsomeclustergceprovider_google_application_credentials_value --value "$(cat $GOOGLE_APPLICATION_CREDENTIALS)" `,
		}
	)

	flags := cmd.Flags()
	flags.StringVar(&value, "value", "", "Secret phrase value used for encrypting and decrypting secrets")
	flags.StringVarP(&file, "file", "s", "", "Secret phrase file used for encrypting and decrypting secrets")
	flags.BoolVar(&stdin, "stdin", false, "Read Secret phrase used for encrypting and decrypting secrets from standard input")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {

		s, err := key(value, file, stdin)
		if err != nil {
			return err
		}

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

		secsMap := make(map[string]interface{})
		if err := yaml.Unmarshal(sec, &secsMap); err != nil {
			panic(err)
		}

		if err := secret.New(logger, secsMap, args[0], orb.Masterkey).Write([]byte(s)); err != nil {
			panic(err)
		}

		if _, err := gitClient.UpdateRemoteUntilItWorks(&git.File{
			Path: "secrets.yml",
			Overwrite: func(o []byte) ([]byte, error) {
				newSecsMap := make(map[string]interface{})
				if err := yaml.Unmarshal(o, &newSecsMap); err != nil {
					panic(err)
				}
				newSecsMap[args[0]] = secsMap[args[0]]
				return orbiter.Marshal(newSecsMap)
			},
			Force: true,
		}); err != nil {
			panic(err)
		}
		return nil
	}
	return cmd
}

func key(value string, file string, stdin bool) (string, error) {

	channels := 0
	if value != "" {
		channels++
	}
	if file != "" {
		channels++
	}
	if stdin {
		channels++
	}

	if channels != 1 {
		return "", errors.New("Key must be provided eighter by value or by file path or by standard input")
	}

	if value != "" {
		return value, nil
	}

	readFunc := func() ([]byte, error) {
		return ioutil.ReadFile(file)
	}
	if stdin {
		readFunc = func() ([]byte, error) {
			return ioutil.ReadAll(os.Stdin)
		}
	}

	key, err := readFunc()
	if err != nil {
		panic(err)
	}
	return string(key), err
}
