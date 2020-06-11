package v1beta1

import "github.com/caos/orbos/internal/operator/boom/api/v1beta1/storage"

type Prometheus struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy" yaml:"deploy"`
	//Spec to define which metrics should get scraped
	//@default: nil
	Metrics *Metrics `json:"metrics,omitempty" yaml:"metrics,omitempty"`
	//Spec to define how the persistence should be handled
	//@default: nil
	Storage *storage.Spec `json:"storage,omitempty" yaml:"storage,omitempty"`
	//Configuration to write to remote prometheus
	RemoteWrite *RemoteWrite `json:"remoteWrite,omitempty" yaml:"remoteWrite,omitempty"`
}

// Metrics: When the metrics spec is nil all metrics will get scraped.
type Metrics struct {
	//Bool if metrics should get scraped from ambassador
	Ambassador bool `json:"ambassador"`
	//Bool if metrics should get scraped from argo-cd
	Argocd bool `json:"argocd"`
	//Bool if metrics should get scraped from kube-state-metrics
	KubeStateMetrics bool `json:"kube-state-metrics" yaml:"kube-state-metrics"`
	//Bool if metrics should get scraped from prometheus-node-exporter
	PrometheusNodeExporter bool `json:"prometheus-node-exporter" yaml:"prometheus-node-exporter"`
	//Bool if metrics should get scraped from prometheus-systemd-exporter
	PrometheusSystemdExporter bool `json:"prometheus-systemd-exporter" yaml:"prometheus-systemd-exporter"`
	//Bool if metrics should get scraped from kube-api-server
	APIServer bool `json:"api-server" yaml:"api-server"`
	//Bool if metrics should get scraped from prometheus-operator
	PrometheusOperator bool `json:"prometheus-operator" yaml:"prometheus-operator"`
	//Bool if metrics should get scraped from logging-operator
	LoggingOperator bool `json:"logging-operator" yaml:"logging-operator"`
	//Bool if metrics should get scraped from loki
	Loki bool `json:"loki"`
	//Bool if metrics should get scraped from boom
	Boom bool `json:"boom" yaml:"boom"`
	//Bool if metrics should get scraped from orbiter
	Orbiter bool `json:"orbiter" yaml:"orbiter"`
}

type RemoteWrite struct {
	//URL of the endpoint of the remote prometheus
	URL string `json:"url" yaml:"url"`
	//Basic-auth-configuration to push metrics to remote prometheus
	BasicAuth *BasicAuth `json:"basicAuth,omitempty" yaml:"basicAuth,omitempty"`
}
type BasicAuth struct {
	//Username to push metrics to remote prometheus
	Username *SecretKeySelector `json:"username" yaml:"username"`
	//Password to push metrics to remote prometheus
	Password *SecretKeySelector `json:"password" yaml:"password"`
}
type SecretKeySelector struct {
	//Name of the existing secret
	Name string `json:"name" yaml:"name"`
	//Name of the key with the value in the existing secret
	Key string `json:"key" yaml:"key"`
}
