package orbiter

import (
	"github.com/caos/orbiter/internal/git"
	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/mntr"
	"gopkg.in/yaml.v3"
)

type AdaptFunc func(monitor mntr.Monitor, desired *Tree, current *Tree) (QueryFunc, DestroyFunc, map[string]*Secret, bool, error)

func parse(gitClient *git.Client, files ...string) (trees []*Tree, err error) {

	if err := gitClient.Clone(); err != nil {
		panic(err)
	}

	for _, file := range files {
		raw, err := gitClient.Read(file)
		if err != nil {
			return nil, err
		}

		tree := &Tree{}
		if err := yaml.Unmarshal([]byte(raw), tree); err != nil {
			return nil, err
		}
		trees = append(trees, tree)
	}

	return trees, nil
}

type Tree struct {
	Common   *Common `yaml:",inline"`
	Original *yaml.Node
	Parsed   interface{} `yaml:",inline"`
}

type Common struct {
	Kind    string
	Version string
}

func (c *Tree) UnmarshalYAML(node *yaml.Node) error {
	c.Original = node
	err := node.Decode(&c.Common)
	return err
}

func (c *Tree) MarshalYAML() (interface{}, error) {
	return c.Parsed, nil
}

type PushSecretsFunc func(monitor mntr.Monitor) error

func pushSecretsFunc(gitClient *git.Client, desired *Tree) PushSecretsFunc {
	return func(monitor mntr.Monitor) error {
		monitor.Info("Writing secret")
		return pushOrbiterYML(monitor, "Secret written", gitClient, desired)
	}
}

func pushOrbiterYML(monitor mntr.Monitor, msg string, gitClient *git.Client, desired *Tree) (err error) {
	monitor.OnChange = func(_ string, fields map[string]string) {
		err = gitClient.UpdateRemote(mntr.SprintCommit(msg, fields), git.File{
			Path:    "orbiter.yml",
			Content: common.MarshalYAML(desired),
		})
		mntr.LogMessage(msg, fields)
	}
	monitor.Changed(msg)
	return err
}
