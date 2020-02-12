package main

import (
	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/caos/orbiter/internal/executables"
	"github.com/caos/orbiter/internal/operator"
	"github.com/caos/orbiter/internal/operator/orbiter"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/orb"
	"github.com/caos/orbiter/internal/watcher/cron"
	"github.com/caos/orbiter/internal/watcher/immediate"
)

func takeoffCommand(rv rootValues) *cobra.Command {

	var (
		verbose bool
		recur   bool
		destroy bool
		deploy  bool
		cmd     = &cobra.Command{
			Use:   "takeoff",
			Short: "Launch an orbiter",
			Long:  "Ensures a desired state",
		}
	)

	flags := cmd.Flags()
	flags.BoolVar(&recur, "recur", false, "Ensure the desired state continously")
	flags.BoolVar(&deploy, "deploy", true, "Ensure Orbiter and Boom deployments continously")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if recur && destroy {
			return errors.New("flags --recur and --destroy are mutually exclusive, please provide eighter one or none")
		}

		ctx, logger, gitClient, orbFile, errFunc := rv()
		if errFunc != nil {
			return errFunc(cmd)
		}

		logger.WithFields(map[string]interface{}{
			"version": version,
			"commit":  gitCommit,
			"destroy": destroy,
			"verbose": verbose,
			"repoURL": orbFile.URL,
		}).Info(false, "Orbiter took off")

		op := operator.New(ctx, logger, orbiter.Takeoff(
			ctx,
			logger,
			gitClient,
			gitCommit,
			orbFile.Masterkey,
			recur,
			orb.AdaptFunc(
				logger,
				orbFile,
				gitCommit,
				!recur,
				deploy),
		), []operator.Watcher{
			immediate.New(logger),
			cron.New(logger, "@every 10s"),
		})

		if err := op.Initialize(); err != nil {
			panic(err)
		}

		executables.Populate()

		op.Run()

		return nil
	}
	return cmd
}
