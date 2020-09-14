package v1beta1

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/argocd"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1/grafana"
)

type Metadata struct {
	//Name of the overlay
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	//Namespace as information, has currently no influence for the applied resources
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

// ToolsetSpec: BOOM reconciles itself if a boomVersion is defined, if no boomVersion is defined there is no reconciling.
type ToolsetSpec struct {
	//Version of BOOM which should be reconciled
	BoomVersion string `json:"boomVersion,omitempty" yaml:"boomVersion,omitempty"`
	//Relative folder path where the currentstate is written to
	ForceApply bool `json:"forceApply,omitempty" yaml:"forceApply,omitempty"`
	//Flag if --force should be used by apply of resources
	CurrentStateFolder string `json:"currentStatePath,omitempty" yaml:"currentStatePath,omitempty"`
	//Spec for the yaml-files applied before the applications, for example used secrets
	PreApply *Apply `json:"preApply,omitempty" yaml:"preApply,omitempty"`
	//Spec for the yaml-files applied after the applications, for example additional crds for the applications
	PostApply *Apply `json:"postApply,omitempty" yaml:"postApply,omitempty"`
	//Spec for the Prometheus-Operator
	PrometheusOperator *PrometheusOperator `json:"prometheus-operator,omitempty" yaml:"prometheus-operator,omitempty"`
	//Spec for the Banzaicloud Logging-Operator
	LoggingOperator *LoggingOperator `json:"logging-operator,omitempty" yaml:"logging-operator,omitempty"`
	//Spec for the Prometheus-Node-Exporter
	PrometheusNodeExporter *PrometheusNodeExporter `json:"prometheus-node-exporter,omitempty" yaml:"prometheus-node-exporter,omitempty"`
	//Spec for the Prometheus-Systemd-Exporter
	PrometheusSystemdExporter *PrometheusSystemdExporter `json:"prometheus-systemd-exporter,omitempty" yaml:"prometheus-systemd-exporter,omitempty"`
	//Spec for the Grafana
	Grafana *grafana.Grafana `json:"grafana,omitempty" yaml:"grafana,omitempty"`
	//Spec for the Ambassador
	Ambassador *Ambassador `json:"ambassador,omitempty" yaml:"ambassador,omitempty"`
	//Spec for the Kube-State-Metrics
	KubeStateMetrics *KubeStateMetrics `json:"kube-state-metrics,omitempty" yaml:"kube-state-metrics,omitempty"`
	//Spec for the Argo-CD
	Argocd *argocd.Argocd `json:"argocd,omitempty" yaml:"argocd,omitempty"`
	//Spec for the Prometheus instance
	Prometheus *Prometheus `json:"prometheus,omitempty" yaml:"prometheus,omitempty"`
	//Spec for the Loki instance
	Loki *Loki `json:"loki,omitempty" yaml:"loki,omitempty"`
	//Spec for Metrics-Server
	MetricsServer *MetricsServer `json:"metrics-server,omitempty" yaml:"metrics-server,omitempty"`
}

type Toolset struct {
	//Version of the used API
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	//Kind for the standard CRD
	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`
	//Metadata for the CRD
	Metadata *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	//Specification for the Toolset
	Spec *ToolsetSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
}

type ToolsetMetadata struct {
	APIVersion string    `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Kind       string    `json:"kind,omitempty" yaml:"kind,omitempty"`
	Metadata   *Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}
