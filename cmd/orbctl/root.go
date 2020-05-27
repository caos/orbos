package main

import (
	"context"
	"github.com/caos/orbos/mntr"
	"github.com/spf13/cobra"
)

type RootValues func() (context.Context, mntr.Monitor, string, errFunc)

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

	return cmd, func() (context.Context, mntr.Monitor, string, errFunc) {

		monitor := mntr.Monitor{
			OnInfo:   mntr.LogMessage,
			OnChange: mntr.LogMessage,
			OnError:  mntr.LogError,
		}

		if verbose {
			monitor = monitor.Verbose()
		}

		return context.Background(), monitor, orbconfig, nil
	}
}
