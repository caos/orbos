package main

import (
	"io/ioutil"
	"os"

	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/orb"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func writeSecretCommand(rv rootValues) *cobra.Command {

	var (
		value string
		file  string
		stdin bool
		cmd   = &cobra.Command{
			Use:   "writesecret [name]",
			Short: "Encrypt and push",
			Args:  cobra.MaximumNArgs(1),
			Example: `orbctl writesecret mystaticprovider.bootstrapkey --file ~/.ssh/my-orb-bootstrap
orbctl writesecret mystaticprovider.bootstrapkey_pub --file ~/.ssh/my-orb-bootstrap.pub
orbctl writesecret mygceprovider.google_application_credentials_value --value "$(cat $GOOGLE_APPLICATION_CREDENTIALS)" `,
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

		_, logger, gitClient, orbconfig, errFunc := rv()
		if errFunc != nil {
			return errFunc(cmd)
		}

		path := ""
		if len(args) > 0 {
			path = args[0]
		}

		if err := orbiter.WriteSecret(
			gitClient,
			orb.AdaptFunc(logger,
				orbconfig,
				gitCommit,
				false,
				false),
			path,
			s); err != nil {
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
