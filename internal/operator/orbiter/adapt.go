package orbiter

import (
	"github.com/caos/orbiter/internal/git"
	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/logging"
	"github.com/caos/orbiter/logging/format"
	"gopkg.in/yaml.v3"
)

type AdaptFunc func(logger logging.Logger, desired *Tree, current *Tree) (EnsureFunc, DestroyFunc, map[string]*Secret, bool, error)

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

type PushSecretsFunc func(logging.Logger) error

func pushSecretsFunc(gitClient *git.Client, desired *Tree) PushSecretsFunc {
	return func(logger logging.Logger) error {
		logger.Info(false, "Writing secret")
		return pushOrbiterYML(logger, "Secret written", gitClient, desired)
	}
}

func pushOrbiterYML(logger logging.Logger, msg string, gitClient *git.Client, desired *Tree) (err error) {
	logger.AddSideEffect(func(event bool, fields map[string]string) {
		err = gitClient.UpdateRemote(format.CommitRecord(fields), git.File{
			Path:    "orbiter.yml",
			Content: common.MarshalYAML(desired),
		})
	}).Info(false, msg)
	return err
}
