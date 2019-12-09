package main

import (
	"strings"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/edge/executables"
	"github.com/caos/orbiter/internal/edge/watcher/cron"
	"github.com/caos/orbiter/internal/edge/watcher/immediate"
	"github.com/caos/orbiter/internal/kinds/orbiter"
	"github.com/caos/orbiter/internal/kinds/orbiter/adapter"
	"github.com/caos/orbiter/internal/kinds/orbiter/model"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"
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
	flags.BoolVar(&verbose, "verbose", false, "Print debug levelled logs")
	flags.BoolVar(&recur, "recur", false, "Ensure the desired state continously")
	flags.BoolVar(&destroy, "destroy", false, "Destroy everything and clean up")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if recur && destroy {
			return errors.New("flags --recur and --destroy are mutually exclusive, please provide eighter one or none")
		}

		ctx, logger, gitClient, repoURL, repokey, masterkey, err := rv(verbose)
		if err != nil {
			return err
		}

		logger.WithFields(map[string]interface{}{
			"version": gitTag,
			"commit":  gitCommit,
			"destroy": destroy,
			"verbose": verbose,
			"repoURL": repoURL,
		}).Info("Orbiter is taking off")

		currentFile := "current.yml"
		secretsFile := "secrets.yml"
		configID := strings.ReplaceAll(strings.TrimSuffix(repoURL[strings.LastIndex(repoURL, "/")+1:], ".git"), "-", "")

		op := operator.New(&operator.Arguments{
			Ctx:         ctx,
			Logger:      logger,
			GitClient:   gitClient,
			MasterKey:   masterkey,
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
				OrbiterVersion:   gitTag,
				NodeagentRepoURL: repoURL,
				NodeagentRepoKey: repokey,
				CurrentFile:      currentFile,
				SecretsFile:      secretsFile,
				Masterkey:        masterkey,
			})),
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
