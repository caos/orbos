package v1beta1

import "github.com/caos/orbos/internal/operator/boom/api/v1beta1/storage"

type Prometheus struct {
	Deploy  bool          `json:"deploy,omitempty" yaml:"deploy,omitempty"`
	Metrics *Metrics      `json:"metrics,omitempty" yaml:"metrics,omitempty"`
	Storage *storage.Spec `json:"storage,omitempty" yaml:"storage,omitempty"`
}

type Metrics struct {
	Ambassador             bool `json:"ambassador"`
	Argocd                 bool `json:"argocd"`
	KubeStateMetrics       bool `json:"kube-state-metrics" yaml:"kube-state-metrics"`
	PrometheusNodeExporter bool `json:"prometheus-node-exporter" yaml:"prometheus-node-exporter"`
	APIServer              bool `json:"api-server" yaml:"api-server"`
	PrometheusOperator     bool `json:"prometheus-operator" yaml:"prometheus-operator"`
	LoggingOperator        bool `json:"logging-operator" yaml:"logging-operator"`
	Loki                   bool `json:"loki"`
}
