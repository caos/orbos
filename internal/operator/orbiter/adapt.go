package orbiter

import (
	"github.com/caos/orbiter/internal/git"
	"github.com/caos/orbiter/internal/operator/common"
	"gopkg.in/yaml.v3"
)

type AdaptFunc func(desired *Tree, secrets *Tree, current *Tree) (EnsureFunc, DestroyFunc, ReadSecretFunc, WriteSecretFunc, error)

func parse(gitClient *git.Client) (desired *Tree, secrets *Tree, err error) {

	if err := gitClient.Clone(); err != nil {
		panic(err)
	}

	rawDesired, err := gitClient.Read("orbiter.yml")
	if err != nil {
		return nil, nil, err
	}
	treeDesired := &Tree{}
	if err := yaml.Unmarshal([]byte(rawDesired), treeDesired); err != nil {
		return nil, nil, err
	}

	rawSecrets, err := gitClient.Read("secrets.yml")
	if err != nil {
		return nil, nil, err
	}

	treeSecrets := &Tree{}
	if err := yaml.Unmarshal([]byte(rawSecrets), treeSecrets); err != nil {
		return nil, nil, err
	}

	return treeDesired, treeSecrets, nil
}

type Tree struct {
	Common   *Common `yaml:",inline"`
	Original yaml.Node
	Parsed   interface{} `yaml:",inline"`
}

type Common struct {
	Kind    string
	Version string
}

func (c *Tree) UnmarshalYAML(node *yaml.Node) error {
	c.Original = *node
	err := node.Decode(&c.Common)
	return err
}

func (c *Tree) MarshalYAML() (interface{}, error) {
	return c.Parsed, nil
}

type PushSecretsFunc func() error

func pushSecretsFunc(gitClient *git.Client, secrets *Tree) PushSecretsFunc {
	return func() error {
		return gitClient.UpdateRemote(git.File{
			Path:    "secrets.yml",
			Content: common.MarshalYAML(secrets),
		})
	}
}
