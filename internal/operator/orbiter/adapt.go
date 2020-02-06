package orbiter

import (
	"github.com/caos/orbiter/internal/git"
	"github.com/caos/orbiter/internal/operator/common"
	"gopkg.in/yaml.v3"
)

type AdaptFunc func(desired *Tree, current *Tree) (EnsureFunc, DestroyFunc, map[string]*Secret, bool, error)

func parse(gitClient *git.Client) (desired *Tree, err error) {

	if err := gitClient.Clone(); err != nil {
		panic(err)
	}

	rawDesired, err := gitClient.Read("orbiter.yml")
	if err != nil {
		return nil, err
	}
	treeDesired := &Tree{}
	if err := yaml.Unmarshal([]byte(rawDesired), treeDesired); err != nil {
		return nil, err
	}

	return treeDesired, nil
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

type PushSecretsFunc func() error

func pushSecretsFunc(gitClient *git.Client, desired *Tree) PushSecretsFunc {
	return func() error {
		return gitClient.UpdateRemote(git.File{
			Path:    "orbiter.yml",
			Content: common.MarshalYAML(desired),
		})
	}
}
