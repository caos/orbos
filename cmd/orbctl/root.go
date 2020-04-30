package main

import (
	"context"
	"github.com/caos/orbos/internal/orb"
	"io/ioutil"

	"github.com/caos/orbos/internal/git"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
)

type RootValues func() (context.Context, mntr.Monitor, *git.Client, *orb.Orb, errFunc)

type errFunc func(cmd *cobra.Command) error

func curryErrFunc(rootCmd *cobra.Command, err error) errFunc {
	return func(cmd *cobra.Command) error {
		cmd.SetUsageFunc(func(_ *cobra.Command) error {
			return rootCmd.Usage()
		})
		return err
	}
}

func RootCommand() (*cobra.Command, RootValues) {

	var (
		verbose   bool
		orbconfig string
	)

	cmd := &cobra.Command{
		Use:   "orbctl [flags]",
		Short: "Interact with your orbs",
		Long: `orbctl launches orbiters and simplifies common tasks such as updating your kubeconfig.
Participate in our community on https://github.com/caos/orbos
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

	return cmd, func() (context.Context, mntr.Monitor, *git.Client, *orb.Orb, errFunc) {

		monitor := mntr.Monitor{
			OnInfo:   mntr.LogMessage,
			OnChange: mntr.LogMessage,
			OnError:  mntr.LogError,
		}

		if verbose {
			monitor = monitor.Verbose()
		}

		content, err := ioutil.ReadFile(orbconfig)
		if err != nil {
			return nil, monitor, nil, nil, curryErrFunc(cmd, err)
		}

		orbStruct := &orb.Orb{}
		if err := yaml.Unmarshal(content, orbStruct); err != nil {
			return nil, monitor, nil, nil, curryErrFunc(cmd, err)
		}

		if orbStruct.URL == "" {
			return nil, monitor, nil, nil, curryErrFunc(cmd, errors.New("orbconfig has no URL configured"))
		}

		if orbStruct.Repokey == "" {
			return nil, monitor, nil, nil, curryErrFunc(cmd, errors.New("orbconfig has no repokey configured"))
		}

		if orbStruct.Masterkey == "" {
			return nil, monitor, nil, nil, curryErrFunc(cmd, errors.New("orbconfig has no masterkey configured"))
		}

		ctx := context.Background()

		gitClient := git.New(ctx, monitor, "Orbiter", "orbiter@caos.ch", orbStruct.URL)
		if err := gitClient.Init([]byte(orbStruct.Repokey)); err != nil {
			panic(err)
		}

		return ctx, monitor, gitClient, orbStruct, nil
	}
}
