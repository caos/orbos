package main

import (
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/edge/executables"
	"github.com/caos/orbiter/internal/edge/watcher/cron"
	"github.com/caos/orbiter/internal/edge/watcher/immediate"
	"github.com/caos/orbiter/internal/kinds/orbiter"
	"github.com/caos/orbiter/internal/kinds/orbiter/adapter"
	"github.com/caos/orbiter/internal/kinds/orbiter/model"
)

func takeoffCommand(rv rootValues) *cobra.Command {

	var (
		verbose bool
		recur   bool
		destroy bool
		cmd     = &cobra.Command{
			Use:   "takeoff",
			Short: "Launch an orbiter",
			Long:  "Ensures a desired state",
		}
	)

	flags := cmd.Flags()
	flags.BoolVar(&recur, "recur", false, "Ensure the desired state continously")
	flags.BoolVar(&destroy, "destroy", false, "Destroy everything and clean up")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if recur && destroy {
			return errors.New("flags --recur and --destroy are mutually exclusive, please provide eighter one or none")
		}

		ctx, logger, gitClient, orb, errFunc := rv()
		if errFunc != nil {
			return errFunc(cmd)
		}

		logger.WithFields(map[string]interface{}{
			"version": version,
			"commit":  gitCommit,
			"destroy": destroy,
			"verbose": verbose,
			"repoURL": orb.URL,
		}).Info("Orbiter is taking off")

		currentFile := "current.yml"
		secretsFile := "secrets.yml"
		configID := strings.ReplaceAll(strings.TrimSuffix(orb.URL[strings.LastIndex(orb.URL, "/")+1:], ".git"), "-", "")

		var before func(desired []byte, secrets *operator.Secrets) error

		if !destroy {
			before = func(desired []byte, secrets *operator.Secrets) error {
				var deserialized struct {
					Spec struct {
						Orbiter string
						Boom    string
					}
					Deps map[string]struct {
						Kind string
					}
				}

				if err := yaml.Unmarshal(desired, &deserialized); err != nil {
					return err
				}

				for clusterName, cluster := range deserialized.Deps {
					if strings.Contains(cluster.Kind, "Kubernetes") {
						if err := ensureArtifacts(logger, secrets, orb, !recur, configID+clusterName, deserialized.Spec.Orbiter, deserialized.Spec.Boom); err != nil {
							return err
						}
					}
				}
				return nil
			}
		}

		op := operator.New(&operator.Arguments{
			Ctx:         ctx,
			Logger:      logger,
			GitClient:   gitClient,
			MasterKey:   orb.Masterkey,
			DesiredFile: "desired.yml",
			CurrentFile: currentFile,
			SecretsFile: secretsFile,
			Watchers: []operator.Watcher{
				immediate.New(logger),
				cron.New(logger, "@every 30s"),
			},
			RootAssembler: orbiter.New(nil, nil, adapter.New(&model.Config{
				Logger:           logger,
				ConfigID:         configID,
				OrbiterVersion:   version,
				OrbiterCommit:    gitCommit,
				NodeagentRepoURL: orb.URL,
				NodeagentRepoKey: orb.Repokey,
				CurrentFile:      currentFile,
				SecretsFile:      secretsFile,
				Masterkey:        orb.Masterkey,
			})),
			BeforeIteration: before,
		})

		iterations := make(chan *operator.IterationDone)
		if err := op.Initialize(); err != nil {
			panic(err)
		}

		executables.Populate()

		go op.Run(iterations)

	outer:
		for it := range iterations {
			if destroy {
				if it.Error != nil {
					panic(it.Error)
				}
				return nil
			}

			if recur {
				if it.Error != nil {
					logger.Error(it.Error)
				}
				continue
			}

			if it.Error != nil {
				panic(it.Error)
			}
			statusReader := struct {
				Deps map[string]struct {
					Current struct {
						State struct {
							Status string
						}
					}
				}
			}{}
			yaml.Unmarshal(it.Current, &statusReader)
			for _, cluster := range statusReader.Deps {
				if cluster.Current.State.Status != "running" {
					continue outer
				}
			}
			break
		}
		return nil
	}
	return cmd
}
