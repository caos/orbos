package v1beta2

import "github.com/caos/orbos/internal/operator/boom/api/v1beta2/storage"

type MetricsPersisting struct {
	Deploy      bool          `json:"deploy" yaml:"deploy"`
	Metrics     *Metrics      `json:"metrics,omitempty" yaml:"metrics,omitempty"`
	Storage     *storage.Spec `json:"storage,omitempty" yaml:"storage,omitempty"`
	RemoteWrite *RemoteWrite  `json:"remoteWrite,omitempty" yaml:"remoteWrite,omitempty"`
}

type Metrics struct {
	Ambassador                bool `json:"ambassador"`
	Argocd                    bool `json:"argocd"`
	KubeStateMetrics          bool `json:"kube-state-metrics" yaml:"kube-state-metrics"`
	PrometheusNodeExporter    bool `json:"prometheus-node-exporter" yaml:"prometheus-node-exporter"`
	PrometheusSystemdExporter bool `json:"prometheus-systemd-exporter" yaml:"prometheus-systemd-exporter"`
	APIServer                 bool `json:"api-server" yaml:"api-server"`
	PrometheusOperator        bool `json:"prometheus-operator" yaml:"prometheus-operator"`
	LoggingOperator           bool `json:"logging-operator" yaml:"logging-operator"`
	Loki                      bool `json:"loki"`
	Boom                      bool `json:"boom" yaml:"boom"`
	Orbiter                   bool `json:"orbiter" yaml:"orbiter"`
}

type RemoteWrite struct {
	URL       string     `json:"url" yaml:"url"`
	BasicAuth *BasicAuth `json:"basicAuth,omitempty" yaml:"basicAuth,omitempty"`
}
type BasicAuth struct {
	Username *SecretKeySelector `json:"username" yaml:"username"`
	Password *SecretKeySelector `json:"password" yaml:"password"`
}
type SecretKeySelector struct {
	Name string `json:"name" yaml:"name"`
	Key  string `json:"key" yaml:"key"`
}
