package main

import (
	_ "net/http/pprof"

	"github.com/caos/orbos/v5/cmd/orbctl/cmds"
	"github.com/spf13/cobra"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

func TakeoffCommand(getRv GetRootValues) *cobra.Command {
	var (
		verbose bool
		recur   bool
		deploy  bool
		cmd     = &cobra.Command{
			Use:   "takeoff",
			Short: "Launch an operator",
			Long:  `For launching specific operators only, pass one or many of "orbiter", "boom" or "networking"`,
			Args:  cobra.MaximumNArgs(3),
		}
	)

	flags := cmd.Flags()
	flags.BoolVar(&recur, "recur", false, "Ensure the desired state continously")
	flags.BoolVar(&deploy, "deploy", true, "Ensure Orbiter and Boom deployments continously")

	cmd.RunE = func(cmd *cobra.Command, args []string) (err error) {

		rv, err := getRv("takeoff", "", map[string]interface{}{"recur": recur, "deploy": deploy, "args": args})
		if err != nil {
			return err
		}
		defer rv.ErrFunc(err)

		orbConfig := rv.OrbConfig
		gitClient := rv.GitClient
		ctx := rv.Ctx

		return cmds.Takeoff(
			monitor,
			ctx,
			orbConfig,
			gitClient,
			recur,
			deploy,
			verbose,
			version,
			gitCommit,
			rv.Kubeconfig,
			rv.Gitops,
			args,
		)
	}
	return cmd
}
