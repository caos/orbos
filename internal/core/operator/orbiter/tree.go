package orbiter

import "gopkg.in/yaml.v3"

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
