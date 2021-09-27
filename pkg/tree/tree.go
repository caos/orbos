package tree

import (
	"fmt"

	"github.com/caos/orbos/v5/mntr"

	"gopkg.in/yaml.v3"
)

var (
	_ yaml.Marshaler   = (*Tree)(nil)
	_ yaml.Unmarshaler = (*Tree)(nil)
)

type Tree struct {
	Common   *Common `yaml:",inline"`
	Original *yaml.Node
	Parsed   interface{} `yaml:",inline"`
}

type Common struct {
	Kind string `json:"kind" yaml:"kind"`
	// Don't access X_Version, it is only exported for (de-)serialization. Use Version and OverwriteVersion methods instead.
	X_Version string `json:"version,omitempty" yaml:"version,omitempty"`
	// Don't access X_ApiVersion, it is only exported for (de-)serialization. Use Version and OverwriteVersion methods instead.
	X_ApiVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
}

func NewCommon(kind, version string, isKubernetesResource bool) *Common {
	c := new(Common)
	c.Kind = kind
	if isKubernetesResource {
		c.X_ApiVersion = version
		return c
	}
	c.X_Version = version
	return c
}

func (c *Common) OverwriteVersion(v string) {
	if c.X_ApiVersion != "" {
		c.X_ApiVersion = v
		return
	}
	c.X_Version = v
}

func (c *Common) Version() string {
	if c.X_ApiVersion != "" {
		return c.X_ApiVersion
	}
	return c.X_Version
}

func (c *Tree) UnmarshalYAML(node *yaml.Node) error {
	c.Original = new(yaml.Node)
	*c.Original = *node

	if err := c.Original.Decode(&c.Common); err != nil {
		return mntr.ToUserError(fmt.Errorf("decoding version or kind failed: kind \"%s\", err %w", c.Common.Kind, err))
	}

	return nil
}

func (c *Tree) MarshalYAML() (interface{}, error) {
	return c.Parsed, nil
}
