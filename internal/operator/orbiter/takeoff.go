package orbiter

import (
	"context"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/caos/orbiter/internal/git"
	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/logging"
)

type EnsureFunc func(psf PushSecretsFunc, nodeAgentsCurrent map[string]*common.NodeAgentCurrent, nodeAgentsDesired map[string]*common.NodeAgentSpec) (err error)

func Takeoff(ctx context.Context, logger logging.Logger, gitClient *git.Client, orbiterCommit string, masterkey string, recur bool, adapt AdaptFunc) func() {

	return func() {

		treeDesired, treeSecrets, err := parse(gitClient)
		if err != nil {
			logger.Error(err)
			return
		}

		treeCurrent := &Tree{}
		ensure, _, _, err := adapt(treeDesired, treeSecrets, treeCurrent)
		if err != nil {
			logger.Error(err)
			return
		}

		if err := gitClient.UpdateRemote(git.File{
			Path:    "orbiter.yml",
			Content: common.MarshalYAML(treeDesired),
		}); err != nil {
			logger.Error(err)
			return
		}

		desiredNodeAgents := make(map[string]*common.NodeAgentSpec)
		currentNodeAgents := common.NodeAgentsCurrentKind{}
		rawCurrentNodeAgents, _ := gitClient.Read("caos-internal/orbiter/node-agents-current.yml")
		yaml.Unmarshal(rawCurrentNodeAgents, &currentNodeAgents)

		if err := ensure(pushSecretsFunc(gitClient, treeSecrets), currentNodeAgents.Current, desiredNodeAgents); err != nil {
			logger.Error(err)
			return
		}

		current := common.MarshalYAML(treeCurrent)

		if err := gitClient.UpdateRemote(git.File{
			Path:    "caos-internal/orbiter/current.yml",
			Content: current,
		}, git.File{
			Path: "caos-internal/orbiter/node-agents-desired.yml",
			Content: common.MarshalYAML(&common.NodeAgentsDesiredKind{
				Kind:    "nodeagent.caos.ch/NodeAgents",
				Version: "v0",
				Spec: common.NodeAgentsSpec{
					Commit:     orbiterCommit,
					NodeAgents: desiredNodeAgents,
				},
			}),
		}); err != nil {
			logger.Error(err)
			return
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
		if err := yaml.Unmarshal(current, &statusReader); err != nil {
			panic(err)
		}
		for _, cluster := range statusReader.Deps {
			if !recur && cluster.Current.State.Status == "running" {
				os.Exit(0)
			}
		}
	}
}
