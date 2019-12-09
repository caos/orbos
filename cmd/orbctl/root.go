package main

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/caos/orbiter/internal/edge/git"
	"github.com/caos/orbiter/logging"
	logcontext "github.com/caos/orbiter/logging/context"
	"github.com/caos/orbiter/logging/stdlib"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

type rootValues func() (context.Context, logging.Logger, *git.Client, string, string, string, error)

var errMultipleStdinKeys = errors.New("Reading multiple keys from standard input does not work")

func rootCommand() (*cobra.Command, rootValues) {

	var (
		verbose        bool
		repoURL        string
		repokey        string
		repokeyFile    string
		repokeyStdin   bool
		masterkey      string
		masterkeyFile  string
		masterkeyStdin bool
	)

	cmd := &cobra.Command{
		Use:   "orbctl",
		Short: "Interact with your orbs",
		Long: `orbctl launches orbiters and simplifies common tasks such as updating your kubeconfig.
Participate in our community on https://github.com/caos/orbiter
or visit our website at https://caos.ch`,
	}

	flags := cmd.PersistentFlags()
	flags.StringVarP(&repoURL, "repourl", "g", "", "Use this orbs Git repo")
	cmd.MarkPersistentFlagRequired("repourl")
	flags.StringVar(&repokey, "repokey", "", "SSH private key value for authenticating to orbs git repo")
	flags.StringVarP(&repokeyFile, "repokey-file", "r", "", "SSH private key file for authenticating to orbs git repo")
	flags.BoolVar(&repokeyStdin, "repokey-stdin", false, "Read SSH private key for authenticating to orbs git repo from standard input")
	flags.StringVar(&masterkey, "masterkey", "", "Secret phrase value used for encrypting and decrypting secrets")
	flags.StringVarP(&masterkeyFile, "masterkey-file", "m", "", "Secret phrase file used for encrypting and decrypting secrets")
	flags.BoolVar(&masterkeyStdin, "masterkey-stdin", false, "Read Secret phrase used for encrypting and decrypting secrets from standard input")
	flags.BoolVar(&verbose, "verbose", false, "Print debug levelled logs")

	return cmd, func() (context.Context, logging.Logger, *git.Client, string, string, string, error) {

		if masterkeyStdin && repokeyStdin {
			return nil, nil, nil, "", "", "", errMultipleStdinKeys
		}

		rk, err := key(repokey, repokeyFile, repokeyStdin)
		if err != nil {
			return nil, nil, nil, "", "", "", errors.Wrap(err, "repokey")
		}

		mk, err := key(masterkey, masterkeyFile, masterkeyStdin)
		if err != nil {
			return nil, nil, nil, "", "", "", errors.Wrap(err, "masterkey")
		}

		ctx := context.Background()

		l := logcontext.Add(stdlib.New(os.Stdout))
		if verbose {
			l = l.Verbose()
		}

		gitClient := git.New(ctx, l, "Orbiter", repoURL)
		if err := gitClient.Init([]byte(rk)); err != nil {
			panic(err)
		}

		return ctx, l, gitClient, repoURL, rk, mk, nil
	}
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
