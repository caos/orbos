package orbiter

import (
	"context"
	"os"

	"gopkg.in/yaml.v3"

	"github.com/caos/orbiter/internal/edge/git"
	"github.com/caos/orbiter/logging"
)

type EnsureFunc func(context.Context) (interface{}, interface{}, error)

type AdaptFunc func(desired *Tree, secrets *Tree, nodeAgentsCurrent map[string]*NodeAgentCurrent) (EnsureFunc, error)

type Common struct {
	Kind    string
	Version string
}

type Tree struct {
	Common   *Common `yaml:",inline"`
	Original *yaml.Node
	Parsed   interface{} `yaml:",inline"`
}

func (c *Tree) UnmarshalYAML(node *yaml.Node) error {
	c.Original = node
	err := node.Decode(&c.Common)
	return err
}

func (c *Tree) MarshalYAML() (interface{}, error) {
	return c.Parsed, nil
}

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

		currentNodeAgents := make(map[string]*NodeAgentCurrent)
		rawCurrentNodeAgents, _ := gitClient.Read("internal/node-agents-current.yml")
		if rawCurrentNodeAgents != nil {
			yaml.Unmarshal(rawCurrentNodeAgents, &currentNodeAgents)
		}

		ensure, err := adapt(treeDesired, treeSecrets, currentNodeAgents)
		if err != nil {
			logger.Error(err)
			return
		}

		treeCurrent, nodeagentsDesired, err := ensure(ctx)
		if err != nil {
			logger.Error(err)
			return
		}

		if _, err := gitClient.UpdateRemoteUntilItWorks(
			&git.File{Path: "desired.yml", Overwrite: func([]byte) ([]byte, error) {
				return yaml.Marshal(treeDesired)
			}}); err != nil {
			panic(err)
		}

		if _, err := gitClient.UpdateRemoteUntilItWorks(
			&git.File{Path: "secrets.yml", Overwrite: func([]byte) ([]byte, error) {
				return yaml.Marshal(treeSecrets)
			}}); err != nil {
			panic(err)
		}

		if _, err := gitClient.UpdateRemoteUntilItWorks(
			&git.File{Path: "internal/node-agents-desired.yml", Overwrite: func([]byte) ([]byte, error) {
				return yaml.Marshal(nodeagentsDesired)
			}}); err != nil {
			panic(err)
		}

		newCurrent, err := gitClient.UpdateRemoteUntilItWorks(
			&git.File{Path: "current.yml", Overwrite: func([]byte) ([]byte, error) {
				return yaml.Marshal(treeCurrent)
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
