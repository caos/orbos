package orbiter

import (
	"context"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/caos/orbiter/internal/git"
	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/logging"
)

type PushSecretsFunc func() error

type EnsureFunc func(psf PushSecretsFunc, nodeAgentsCurrent map[string]*common.NodeAgentCurrent, nodeAgentsDesired map[string]*common.NodeAgentSpec) (err error)

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

		if err := ensure(func() error {
			return gitClient.UpdateRemote(git.File{
				Path:    "secrets.yml",
				Content: common.MarshalYAML(treeSecrets),
			})
		}, currentNodeAgents.Current, desiredNodeAgents); err != nil {
			logger.Error(err)
			return
		}

		current := common.MarshalYAML(treeCurrent)

		if err := gitClient.UpdateRemote(git.File{
			Path:    "desired.yml",
			Content: common.MarshalYAML(treeDesired),
		}, git.File{
			Path:    "current.yml",
			Content: current,
		}, git.File{
			Path: "internal/node-agents-desired.yml",
			Content: common.MarshalYAML(&common.NodeAgentsDesiredKind{
				Kind:    "nodeagent.caos.ch/NodeAgent",
				Version: "v0",
				Spec: common.NodeAgentsSpec{
					Commit:     orbiterCommit,
					NodeAgents: desiredNodeAgents,
				},
			}),
		}); err != nil {
			panic(err)
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
			if destroy && cluster.Current.State.Status == "destroyed" ||
				!destroy && !recur && cluster.Current.State.Status == "running" {
				os.Exit(0)
			}
		}
	}

}
