package orbiter

import (
	"context"
	"os"

	"github.com/pkg/errors"
	"gopkg.in/yaml.v2"

	"github.com/caos/orbiter/internal/edge/git"
	"github.com/caos/orbiter/logging"
)

func Iterator(ctx context.Context, logger logging.Logger, gitClient *git.Client, orbiterCommit string, masterkey string, recur bool, destroy bool, orbAssembler Assembler) func() {

	return func() {
		if err := gitClient.Clone(); err != nil {
			panic(err)
		}

		desiredBytes, err := gitClient.Read("desired.yml")
		if err != nil {
			logger.Error(err)
			return
		}

		desired := make(map[string]interface{})
		if err := yaml.Unmarshal(desiredBytes, &desired); err != nil {
			logger.Error(err)
			return
		}
		/*
			if i.args.BeforeIteration != nil {
				if err := i.args.BeforeIteration(desiredBytes, secrets); err != nil {
					return err
				}
			}
		*/
		secretsBytes, err := gitClient.Read("secrets.yml")
		if err != nil {
			logger.Error(err)
			return
		}

		secretsMap := make(map[string]interface{})
		if err := yaml.Unmarshal(secretsBytes, &secretsMap); err != nil {
			panic(err)
		}

		curriedSecrets := currySecrets(logger, func(newSecrets map[string]interface{}) error {
			_, err := gitClient.UpdateRemoteUntilItWorks(&git.File{
				Path: "secrets.yml",
				Overwrite: func([]byte) ([]byte, error) {
					return Marshal(newSecrets)
				},
				Force: true,
			})
			return err
		}, secretsMap, masterkey)

		secrets := &Secrets{curriedSecrets.read, curriedSecrets.write, curriedSecrets.delete}

		currentNodeAgentBytes, _ := gitClient.Read("internal/node-agents-current.yml")
		tree, err := build(logger, orbAssembler, desired, secrets, nil, true, newNodeAgentCurrentFunc(logger, currentNodeAgentBytes))
		if err != nil {
			logger.Error(err)
			return
		}

		if _, err := ensure(ctx, logger, tree, secrets); err != nil {
			logger.Error(err)
			return
		}

		newCurrent, err := gitClient.UpdateRemoteUntilItWorks(
			&git.File{Path: "current.yml", Overwrite: func(reloadedCurrent []byte) ([]byte, error) {

				curr := make(map[string]interface{})

				if err := buildCurrent(logger, curr, tree, orbiterCommit); err != nil {
					return nil, errors.Wrap(err, "overwriting current state failed")
				}

				return Marshal(curr)
			}})

		if err != nil {
			panic(err)
		}

		statusReader := struct {
			Deps struct {
				Clusters map[string]struct {
					Current struct {
						State struct {
							Status string
						}
					}
				}
			}
		}{}
		if err := yaml.Unmarshal(newCurrent, &statusReader); err != nil {
			panic(err)
		}
		for _, cluster := range statusReader.Deps.Clusters {
			if destroy && cluster.Current.State.Status != "destroyed" ||
				!destroy && !recur && cluster.Current.State.Status == "running" {
				os.Exit(0)
			}
		}
	}

}
