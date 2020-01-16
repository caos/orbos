package orbiter

import (
	"context"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/caos/orbiter/internal/core/operator/common"
	"github.com/caos/orbiter/internal/edge/git"
	"github.com/caos/orbiter/logging"
)

type EnsureFunc func(nodeAgentsCurrent map[string]*common.NodeAgentCurrent, nodeAgentsDesired map[string]*common.NodeAgentSpec) (err error)

type AdaptFunc func(desired *Tree, secrets *Tree, current *Tree) (EnsureFunc, error)

func Iterator(ctx context.Context, logger logging.Logger, gitClient *git.Client, orbiterCommit string, masterkey string, recur bool, destroy bool, adapt AdaptFunc) func() {

	return func() {

		if err := gitClient.Clone(); err != nil {
			panic(err)
		}

		rawDesired, err := gitClient.Read("desired.yml")
		if err != nil {
			logger.Error(err)
			return
		}
		treeDesired := &Tree{}
		if err := yaml.Unmarshal([]byte(rawDesired), treeDesired); err != nil {
			panic(err)
		}

		rawSecrets, err := gitClient.Read("secrets.yml")
		if err != nil {
			logger.Error(err)
			return
		}
		treeSecrets := &Tree{}
		if err := yaml.Unmarshal([]byte(rawSecrets), treeSecrets); err != nil {
			panic(err)
		}

		treeCurrent := &Tree{}
		ensure, err := adapt(treeDesired, treeSecrets, treeCurrent)
		if err != nil {
			logger.Error(err)
			return
		}

		desiredNodeAgents := make(map[string]*common.NodeAgentSpec)
		currentNodeAgents := common.NodeAgentsCurrentKind{}
		rawCurrentNodeAgents, _ := gitClient.Read("internal/node-agents-current.yml")
		yaml.Unmarshal(rawCurrentNodeAgents, &currentNodeAgents)

		if err := ensure(currentNodeAgents.Current, desiredNodeAgents); err != nil {
			logger.Error(err)
			return
		}

		if _, err := gitClient.UpdateRemoteUntilItWorks(
			&git.File{Path: "desired.yml", Overwrite: func([]byte) ([]byte, error) {
				return common.MarshalYAML(treeDesired), nil
			}}); err != nil {
			panic(err)
		}

		if _, err := gitClient.UpdateRemoteUntilItWorks(
			&git.File{Path: "secrets.yml", Overwrite: func([]byte) ([]byte, error) {
				return common.MarshalYAML(treeSecrets), nil
			}}); err != nil {
			panic(err)
		}

		if _, err := gitClient.UpdateRemoteUntilItWorks(
			&git.File{Path: "internal/node-agents-desired.yml", Overwrite: func([]byte) ([]byte, error) {
				return common.MarshalYAML(&common.NodeAgentsDesiredKind{
					Kind:    "nodeagent.caos.ch/NodeAgent",
					Version: "v0",
					Spec: common.NodeAgentsSpec{
						Commit:     orbiterCommit,
						NodeAgents: desiredNodeAgents,
					},
				}), nil
			}}); err != nil {
			panic(err)
		}

		newCurrent, err := gitClient.UpdateRemoteUntilItWorks(
			&git.File{Path: "current.yml", Overwrite: func([]byte) ([]byte, error) {
				return common.MarshalYAML(treeCurrent), nil
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
