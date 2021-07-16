package tree

import (
	"fmt"

	"github.com/caos/orbos/mntr"

	"gopkg.in/yaml.v3"
)

type Tree struct {
	Common   *Common `yaml:",inline"`
	Original *yaml.Node
	Parsed   interface{} `yaml:",inline"`
}

type Common struct {
	Kind    string `json:"kind"`
	Version string `json:"version" yaml:"version" yaml:"apiVersion"`
}

func (c *Tree) UnmarshalYAML(node *yaml.Node) error {
	c.Original = new(yaml.Node)
	*c.Original = *node

	c.Common = new(Common)
	if err := c.Original.Decode(c.Common); err != nil || c.Common.Version == "" || c.Common.Kind == "" {
		return mntr.ToUserError(fmt.Errorf("decoding version or kind failed: kind \"%s\", version \"%s\", err %w", c.Common.Version, c.Common.Kind, err))
	}
	return nil
}

func (c *Tree) MarshalYAML() (interface{}, error) {
	return c.Parsed, nil
}
