package main

import (
	"context"
	"flag"

	"github.com/caos/orbos/internal/helpers"

	"github.com/caos/orbos/pkg/git"

	"github.com/spf13/cobra"

	"github.com/caos/orbos/internal/orb"
	"github.com/caos/orbos/mntr"
)

type RootValues struct {
	Ctx         context.Context
	Monitor     mntr.Monitor
	OrbConfig   *orb.Orb
	GitClient   *git.Client
	MetricsAddr string
	ErrFunc     errFunc
}

type GetRootValues func() (*RootValues, error)

type errFunc func(err error) error

func RootCommand() (*cobra.Command, GetRootValues) {

	var (
		verbose       bool
		orbConfigPath string
		metricsAddr   string
	)

	cmd := &cobra.Command{
		Use:   "orbctl [flags]",
		Short: "Interact with your orbs",
		Long: `orbctl launches orbiters, booms, database-operators and networking-operators and simplifies common tasks such as updating your kubeconfig.
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
	flags.StringVarP(&orbConfigPath, "orbconfig", "f", "~/.orb/config", "Path to the file containing the orbs git repo URL, deploy key and the master key for encrypting and decrypting secrets")
	flags.BoolVar(&verbose, "verbose", false, "Print debug levelled logs")
	flag.StringVar(&metricsAddr, "metrics-addr", "", "The address the metric endpoint binds to.")

	return cmd, func() (*RootValues, error) {

		if verbose {
			monitor = monitor.Verbose()
		}

		prunedPath := helpers.PruneHome(orbConfigPath)
		orbConfig, err := orb.ParseOrbConfig(prunedPath)
		if err != nil {
			orbConfig = &orb.Orb{Path: prunedPath}
			return nil, err
		}

		ctx := context.Background()

		return &RootValues{
			Ctx:         ctx,
			Monitor:     monitor,
			OrbConfig:   orbConfig,
			GitClient:   git.New(ctx, monitor, "orbos", "orbos@caos.ch"),
			MetricsAddr: metricsAddr,
			ErrFunc: func(err error) error {
				if err != nil {
					monitor.Error(err)
				}
				return nil
			},
		}, nil
	}
}
