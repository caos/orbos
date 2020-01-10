package main

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/core/operator/orbiter"
	"github.com/caos/orbiter/internal/edge/executables"
	"github.com/caos/orbiter/internal/edge/watcher/cron"
	"github.com/caos/orbiter/internal/edge/watcher/immediate"
	"github.com/caos/orbiter/internal/kinds/orb"
	"github.com/caos/orbiter/internal/kinds/orb/adapter"
	"github.com/caos/orbiter/internal/kinds/orb/model"
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
	flags.BoolVar(&destroy, "destroy", false, "Destroy everything and clean up")
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
		}).Info("Orbiter is taking off")

		currentFile := "current.yml"
		secretsFile := "secrets.yml"
		configID := strings.ReplaceAll(strings.TrimSuffix(orbFile.URL[strings.LastIndex(orbFile.URL, "/")+1:], ".git"), "-", "")
		/*
			var before func(desired []byte, secrets *operator.Secrets) error

			if deploy && !destroy {
				before = func(desired []byte, secrets *operator.Secrets) error {
					var deserialized struct {
						Spec struct {
							Orbiter string
							Boom    string
							Verbose bool
						}
						Deps map[string]struct {
							Kind string
						}
					}

					if err := yaml.Unmarshal(desired, &deserialized); err != nil {
						return err
					}

					l := logger
					if deserialized.Spec.Verbose {
						l = logger.Verbose()
					}

					for clusterName, cluster := range deserialized.Deps {
						if strings.Contains(cluster.Kind, "Kubernetes") {
							if err := ensureArtifacts(l, secrets, orb, !recur, configID+clusterName, deserialized.Spec.Orbiter, deserialized.Spec.Boom); err != nil {
								return err
							}
						}
					}
					return nil
				}
			}
		*/
		op := operator.New(ctx, logger, orbiter.Iterator(ctx, logger, gitClient, gitCommit, orbFile.Masterkey, recur, destroy, orb.New(nil, nil, adapter.New(&model.Config{
			Logger:             logger,
			ConfigID:           configID,
			OrbiterCommit:      gitCommit,
			NodeagentRepoURL:   orbFile.URL,
			NodeagentRepoKey:   orbFile.Repokey,
			CurrentFile:        currentFile,
			SecretsFile:        secretsFile,
			Masterkey:          orbFile.Masterkey,
			ConnectFromOutside: !recur,
		}))), []operator.Watcher{
			immediate.New(logger),
			cron.New(logger, "@every 10s"),
		})

		if err := op.Initialize(); err != nil {
			panic(err)
		}

		executables.Populate()

		go op.Run()

		return nil
	}
	return cmd
}
