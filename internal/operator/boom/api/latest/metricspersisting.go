package latest

import (
	"github.com/caos/orbos/internal/operator/boom/api/latest/storage"
	"github.com/caos/orbos/pkg/kubernetes/k8s"
)

type MetricsPersisting struct {
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
	//Static labels added to metrics
	ExternalLabels map[string]string `json:"externalLabels,omitempty" yaml:"externalLabels,omitempty"`
	//NodeSelector for statefulset
	NodeSelector map[string]string `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	//Tolerations to run prometheus on nodes
	Tolerations k8s.Tolerations `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`
	//Resource requirements
	Resources *k8s.Resources `json:"resources,omitempty" yaml:"resources,omitempty"`
	//Override used image version
	OverwriteVersion string `json:"overwriteVersion,omitempty" yaml:"overwriteVersion,omitempty"`
}

// Metrics: When the metrics spec is nil all metrics will get scraped.
type Metrics struct {
	//Bool if metrics should get scraped from ambassador
	Ambassador bool `json:"ambassador"`
	//Bool if metrics should get scraped from argo-cd
	Argocd bool `json:"argocd"`
	//Bool if metrics should get scraped from kube-state-metrics
	KubeStateMetrics bool `json:"kube-state-metrics" yaml:"kube-state-metrics"`
	//Bool if metrics should get scraped from prometheus
	Prometheus bool `json:"prometheus" yaml:"prometheus"`
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
	//Bool if metrics should get scraped from BOOM
	Boom bool `json:"boom" yaml:"boom"`
	//Bool if metrics should get scraped from ORBITER
	Orbiter bool `json:"orbiter" yaml:"orbiter"`
	//Bool if metrics should get scraped from ZITADEL
	Zitadel bool `json:"zitadel" yaml:"zitadel"`
	//Bool if metrics should get scraped from database
	Database bool `json:"database" yaml:"database"`
	//Bool if metrics should get scraped from networking
	Networking bool `json:"networking" yaml:"networking"`
}

type RemoteWrite struct {
	//URL of the endpoint of the remote prometheus
	URL string `json:"url" yaml:"url"`
	//Basic-auth-configuration to push metrics to remote prometheus
	BasicAuth *BasicAuth `json:"basicAuth,omitempty" yaml:"basicAuth,omitempty"`
	//RelabelConfigs for remote write
	RelabelConfigs []*RelabelConfig `json:"relabelConfigs,omitempty" yaml:"relabelConfigs,omitempty"`
}
type RelabelConfig struct {
	//The source labels select values from existing labels. Their content is concatenated using the configured separator and matched against the configured regular expression for the replace, keep, and drop actions.
	SourceLabels []string `json:"sourceLabels,omitempty" yaml:"sourceLabels,omitempty"`
	//Separator placed between concatenated source label values. default is ';'.
	Separator string `json:"separator,omitempty" yaml:"separator,omitempty"`
	//Label to which the resulting value is written in a replace action. It is mandatory for replace actions. Regex capture groups are available.
	TargetLabel string `json:"targetLabel,omitempty" yaml:"targetLabel,omitempty"`
	//Regular expression against which the extracted value is matched. Default is '(.*)'
	Regex string `json:"regex,omitempty" yaml:"regex,omitempty"`
	//Modulus to take of the hash of the source label values.
	Modulus string `json:"modulus,omitempty" yaml:"modulus,omitempty"`
	//Replacement value against which a regex replace is performed if the regular expression matches. Regex capture groups are available. Default is '$1'
	Replacement string `json:"replacement,omitempty" yaml:"replacement,omitempty"`
	//Action to perform based on regex matching. Default is 'replace'
	Action string `json:"action,omitempty" yaml:"action,omitempty"`
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
