package v1beta1

type Metadata struct {
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

type ToolsetSpec struct {
	ForceApply             bool                    `json:"forceApply,omitempty" yaml:"forceApply,omitempty"`
	CurrentStateFolder     string                  `json:"currentStatePath,omitempty" yaml:"currentStatePath,omitempty"`
	PreApply               *PreApply               `json:"preApply,omitempty" yaml:"preApply,omitempty"`
	PostApply              *PostApply              `json:"postApply,omitempty" yaml:"postApply,omitempty"`
	PrometheusOperator     *PrometheusOperator     `json:"prometheus-operator,omitempty" yaml:"prometheus-operator"`
	LoggingOperator        *LoggingOperator        `json:"logging-operator,omitempty" yaml:"logging-operator"`
	PrometheusNodeExporter *PrometheusNodeExporter `json:"prometheus-node-exporter,omitempty" yaml:"prometheus-node-exporter"`
	Grafana                *Grafana                `json:"grafana,omitempty" yaml:"grafana"`
	Ambassador             *Ambassador             `json:"ambassador,omitempty" yaml:"ambassador"`
	KubeStateMetrics       *KubeStateMetrics       `json:"kube-state-metrics,omitempty" yaml:"kube-state-metrics"`
	Argocd                 *Argocd                 `json:"argocd,omitempty" yaml:"argocd"`
	Prometheus             *Prometheus             `json:"prometheus,omitempty" yaml:"prometheus"`
	Loki                   *Loki                   `json:"loki,omitempty" yaml:"loki"`
}

type Toolset struct {
	APIVersion string       `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Kind       string       `json:"kind,omitempty" yaml:"kind,omitempty"`
	Metadata   *Metadata    `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	Spec       *ToolsetSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
}

type ToolsetMetadata struct {
	APIVersion string    `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Kind       string    `json:"kind,omitempty" yaml:"kind,omitempty"`
	Metadata   *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}
