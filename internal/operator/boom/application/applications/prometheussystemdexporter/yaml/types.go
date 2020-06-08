package yaml

type DaemonSet struct {
	Kind       string        `yaml:"kind,omitempty"`
	ApiVersion string        `yaml:"apiVersion,omitempty"`
	Metadata   Metadata      `yaml:"metadata,omitempty"`
	Spec       DaemonSetSpec `yaml:"spec,omitempty"`
}

type Metadata struct {
	Name        string            `yaml:"name,omitempty"`
	Namespace   string            `yaml:"namespace,omitempty"`
	Labels      map[string]string `yaml:"labels,omitempty"`
	Annotations map[string]string `yaml:"annotations,omitempty"`
}

type DaemonSetSpec struct {
	Selector       Selector
	UpdateStrategy UpdateStrategy
	Template       PodTemplate
}

type Selector struct {
	MatchLabels map[string]string
}

type UpdateStrategy struct {
	Type          string
	RollingUpdate struct {
		MaxUnavailable string
	}
}

type RollingUpdate struct {
	MaxUnavailable string
}

type PodTemplate struct {
	Metadata Metadata
	Spec     PodSpec
}

type PodSpec struct {
	Tolerations     []*Toleration
	SecurityContext struct{}
	HostPID         bool
	Containers      []struct{}
	Volumes         []struct{}
}

type Toleration struct {
	Key      string
	Operator string
	Effect   string
}
