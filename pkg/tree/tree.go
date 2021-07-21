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
	Kind    string `json:"kind" yaml:"kind"`
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	// Don't access X_ApiVersion, it is only exported for (de-)serialization. Access the Version property only.
	X_ApiVersion     string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	parsedApiVersion bool   `json:"-" yaml:"-"`
}

func (c *Common) UnmarshalYAML(node *yaml.Node) error {
	type proxy Common
	p := proxy{}
	if err := node.Decode(&p); err != nil {
		return err
	}
	c.Kind = p.Kind
	c.Version = p.Version
	c.X_ApiVersion = p.X_ApiVersion
	if c.Version == "" && c.X_ApiVersion != "" {
		c.Version = c.X_ApiVersion
		c.parsedApiVersion = true
	}
	return nil
}

func (c *Common) MarshalYAML() (interface{}, error) {
	type proxy Common
	clone := proxy{
		Kind:             c.Kind,
		Version:          c.Version,
		X_ApiVersion:     c.X_ApiVersion,
		parsedApiVersion: c.parsedApiVersion,
	}
	if c.parsedApiVersion {
		clone.Version = ""
		clone.X_ApiVersion = c.Version
		return clone, nil
	}
	return clone, nil
}

func (c *Tree) UnmarshalYAML(node *yaml.Node) error {
	c.Original = new(yaml.Node)
	*c.Original = *node

	c.Common = new(Common)
	if err := c.Original.Decode(c.Common); err != nil {
		return mntr.ToUserError(fmt.Errorf("decoding version or kind failed: kind \"%s\", version \"%s\", err %w", c.Common.Kind, c.Common.Version, err))
	}

	return nil
}

func (c *Tree) MarshalYAML() (interface{}, error) {
	return c.Parsed, nil
}
