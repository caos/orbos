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
	"gopkg.in/yaml.v2"
)

type Orb struct {
	URL       string
	Repokey   string
	Masterkey string
}

type rootValues func() (context.Context, logging.Logger, *git.Client, *Orb, errFunc)

type errFunc func(cmd *cobra.Command) error

func curryErrFunc(rootCmd *cobra.Command, err error) errFunc {
	return func(cmd *cobra.Command) error {
		cmd.SetUsageFunc(func(_ *cobra.Command) error {
			return rootCmd.Usage()
		})
		return err
	}
}

func rootCommand() (*cobra.Command, rootValues) {

	var (
		verbose   bool
		orbconfig string
	)

	cmd := &cobra.Command{
		Use:   "orbctl [flags]",
		Short: "Interact with your orbs",
		Long: `orbctl launches orbiters and simplifies common tasks such as updating your kubeconfig.
Participate in our community on https://github.com/caos/orbiter
and visit our website at https://caos.ch`,
		Example: `$ mkdir -p ~/.orb
$ cat > ~/.orb/myorb << EOF
> url: git@github.com:me/my-orb.git
> masterkey: "$(gopass my-secrets/orbs/myorb/masterkey)"
> repokey: |
> $(cat ~/.ssh/myorbrepo | sed s/^/\ \ /g)
> EOF
$ orbctl -f ~/.orb/myorb [command]
`,
	}

	flags := cmd.PersistentFlags()
	flags.StringVarP(&orbconfig, "orbconfig", "f", "~/.orb/config", "Path to the file containing the orbs git repo URL, deploy key and the master key for encrypting and decrypting secrets")
	flags.BoolVar(&verbose, "verbose", false, "Print debug levelled logs")

	return cmd, func() (context.Context, logging.Logger, *git.Client, *Orb, errFunc) {

		content, err := ioutil.ReadFile(orbconfig)
		if err != nil {
			return nil, nil, nil, nil, curryErrFunc(cmd, err)
		}

		orb := &Orb{}
		if err := yaml.Unmarshal(content, orb); err != nil {
			return nil, nil, nil, nil, curryErrFunc(cmd, err)
		}

		if orb.URL == "" {
			return nil, nil, nil, nil, curryErrFunc(cmd, errors.New("orbconfig has no URL configured"))
		}

		if orb.Repokey == "" {
			return nil, nil, nil, nil, curryErrFunc(cmd, errors.New("orbconfig has no repokey configured"))
		}

		if orb.Masterkey == "" {
			return nil, nil, nil, nil, curryErrFunc(cmd, errors.New("orbconfig has no masterkey configured"))
		}

		ctx := context.Background()

		l := logcontext.Add(stdlib.New(os.Stdout))
		if verbose {
			l = l.Verbose()
		}

		gitClient := git.New(ctx, l, "Orbiter", orb.URL)
		if err := gitClient.Init([]byte(orb.Repokey)); err != nil {
			panic(err)
		}

		return ctx, l, gitClient, orb, nil
	}
}
