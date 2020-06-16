package common

type Metadata struct {
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

type Toolset struct {
	APIVersion string       `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Kind       string       `json:"kind,omitempty" yaml:"kind,omitempty"`
	Metadata   *Metadata    `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Spec       *interface{} `json:"spec,omitempty" yaml:"spec,omitempty"`
}

func New() *Toolset {
	return &Toolset{}
}
