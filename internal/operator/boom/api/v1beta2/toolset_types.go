package v1beta2

import (
	"github.com/caos/orbos/v5/internal/operator/boom/api/latest"
	"github.com/caos/orbos/v5/internal/operator/boom/api/latest/monitoring"
	"github.com/caos/orbos/v5/internal/operator/boom/api/latest/reconciling"
)

// ToolsetSpec: BOOM reconciles itself if a boomVersion is defined, if no boomVersion is defined there is no reconciling.
type ToolsetSpec struct {
	//Boom self reconciling specs
	Boom *latest.Boom `json:"boom,omitempty" yaml:"boom,omitempty"`
	//Flag if --force should be used by apply of resources
	ForceApply bool `json:"forceApply,omitempty" yaml:"forceApply,omitempty"`
	//Relative folder path where the currentstate is written to
	CurrentStateFolder string `json:"currentStatePath,omitempty" yaml:"currentStatePath,omitempty"`
	//Spec for the yaml-files applied before the applications, for example used secrets
	PreApply *latest.Apply `json:"preApply,omitempty" yaml:"preApply,omitempty"`
	//Spec for the yaml-files applied after the applications, for example additional crds for the applications
	PostApply *latest.Apply `json:"postApply,omitempty" yaml:"postApply,omitempty"`
	//Spec for the Prometheus-Operator
	MetricCollection *latest.MetricCollection `json:"metricCollection,omitempty" yaml:"metricCollection,omitempty"`
	//Spec for the Banzaicloud Logging-Operator
	LogCollection *LogCollection `json:"logCollection,omitempty" yaml:"logCollection,omitempty"`
	//Spec for the Prometheus-Node-Exporter
	NodeMetricsExporter *latest.NodeMetricsExporter `json:"nodeMetricsExporter,omitempty" yaml:"nodeMetricsExporter,omitempty"`
	//Spec for the Prometheus-Systemd-Exporter
	SystemdMetricsExporter *latest.SystemdMetricsExporter `json:"systemdMetricsExporter,omitempty" yaml:"systemdMetricsExporter,omitempty"`
	//Spec for the Grafana
	Monitoring *monitoring.Monitoring `json:"monitoring,omitempty" yaml:"monitoring,omitempty"`
	//Spec for the Ambassador
	APIGateway *latest.APIGateway `json:"apiGateway,omitempty" yaml:"apiGateway,omitempty"`
	//Spec for the Kube-State-Metrics
	KubeMetricsExporter *latest.KubeMetricsExporter `json:"kubeMetricsExporter,omitempty" yaml:"kubeMetricsExporter,omitempty"`
	//Spec for the Argo-CD
	Reconciling *reconciling.Reconciling `json:"reconciling,omitempty" yaml:"reconciling,omitempty"`
	//Spec for the Prometheus instance
	MetricsPersisting *latest.MetricsPersisting `json:"metricsPersisting,omitempty" yaml:"metricsPersisting,omitempty"`
	//Spec for the Loki instance
	LogsPersisting *latest.LogsPersisting `json:"logsPersisting,omitempty" yaml:"logsPersisting,omitempty"`
	//Spec for Metrics-Server
	MetricsServer *latest.MetricsServer `json:"metricsServer,omitempty" yaml:"metricsServer,omitempty"`
}

type Toolset struct {
	//Version of the used API
	APIVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	//Kind for the standard CRD
	Kind string `json:"kind,omitempty" yaml:"kind,omitempty"`
	//Metadata for the CRD
	Metadata *latest.Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
	//Specification for the Toolset
	Spec *ToolsetSpec `json:"spec,omitempty" yaml:"spec,omitempty"`
}

type ToolsetMetadata struct {
	APIVersion string           `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Kind       string           `json:"kind,omitempty" yaml:"kind,omitempty"`
	Metadata   *latest.Metadata `json:"metadata,omitempty" yaml:"metadata,omitempty"`
}
