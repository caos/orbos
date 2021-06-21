package tree

import "gopkg.in/yaml.v3"

type Tree struct {
	Common   *Common `yaml:",inline"`
	Original *yaml.Node
	Parsed   interface{} `yaml:",inline"`
}

type Common struct {
	//Kind of the used Type, which gives the structure for spec
	Kind string `json:"kind"`
	//Version of the used Kind
	Version string `json:"version" yaml:"version" yaml:"apiVersion"`
}

func (c *Tree) UnmarshalYAML(node *yaml.Node) error {
	c.Original = new(yaml.Node)
	*c.Original = *node

	c.Common = new(Common)
	err := c.Original.Decode(c.Common)
	return err
}

func (c *Tree) MarshalYAML() (interface{}, error) {
	return c.Parsed, nil
}
