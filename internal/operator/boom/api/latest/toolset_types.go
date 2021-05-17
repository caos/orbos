package latest

import (
	"github.com/caos/orbos/internal/operator/boom/api/latest/monitoring"
	"github.com/caos/orbos/internal/operator/boom/api/latest/reconciling"
	"github.com/caos/orbos/pkg/kubernetes/k8s"
)

type Metadata struct {
	//Name of the overlay
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	//Namespace as information, has currently no influence for the applied resources
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

// ToolsetSpec: BOOM reconciles itself if a boomVersion is defined, if no boomVersion is defined there is no reconciling.
type ToolsetSpec struct {
	//Boom self reconciling specs
	Boom *Boom `json:"boom,omitempty" yaml:"boom,omitempty"`
	//Flag if --force should be used by apply of resources
	ForceApply bool `json:"forceApply,omitempty" yaml:"forceApply,omitempty"`
	//Relative folder path where the currentstate is written to
	CurrentStateFolder string `json:"currentStatePath,omitempty" yaml:"currentStatePath,omitempty"`
	//Spec for the yaml-files applied before the applications, for example used secrets
	PreApply *Apply `json:"preApply,omitempty" yaml:"preApply,omitempty"`
	//Spec for the yaml-files applied after the applications, for example additional crds for the applications
	PostApply *Apply `json:"postApply,omitempty" yaml:"postApply,omitempty"`
	//Spec for the Prometheus-Operator
	MetricCollection *MetricCollection `json:"metricCollection,omitempty" yaml:"metricCollection,omitempty"`
	//Spec for the Banzaicloud Logging-Operator
	LogCollection *LogCollection `json:"logCollection,omitempty" yaml:"logCollection,omitempty"`
	//Spec for the Prometheus-Node-Exporter
	NodeMetricsExporter *NodeMetricsExporter `json:"nodeMetricsExporter,omitempty" yaml:"nodeMetricsExporter,omitempty"`
	//Spec for the Prometheus-Systemd-Exporter
	SystemdMetricsExporter *SystemdMetricsExporter `json:"systemdMetricsExporter,omitempty" yaml:"systemdMetricsExporter,omitempty"`
	//Spec for the Grafana
	Monitoring *monitoring.Monitoring `json:"monitoring,omitempty" yaml:"monitoring,omitempty"`
	//Spec for the Ambassador
	APIGateway *APIGateway `json:"apiGateway,omitempty" yaml:"apiGateway,omitempty"`
	//Spec for the Kube-State-Metrics
	KubeMetricsExporter *KubeMetricsExporter `json:"kubeMetricsExporter,omitempty" yaml:"kubeMetricsExporter,omitempty"`
	//Spec for the Argo-CD
	Reconciling *reconciling.Reconciling `json:"reconciling,omitempty" yaml:"reconciling,omitempty"`
	//Spec for the Prometheus instance
	MetricsPersisting *MetricsPersisting `json:"metricsPersisting,omitempty" yaml:"metricsPersisting,omitempty"`
	//Spec for the Loki instance
	LogsPersisting *LogsPersisting `json:"logsPersisting,omitempty" yaml:"logsPersisting,omitempty"`
	//Spec for Metrics-Server
	MetricsServer *MetricsServer `json:"metricsServer,omitempty" yaml:"metricsServer,omitempty"`
	//Spec for Minio-Operator
	S3StorageOperator *S3StorageOperator `json:"s3StorageOperator,omitempty" yaml:"s3StorageOperator,omitempty"`
}

type Boom struct {
	//Version of BOOM which should be reconciled
	Version string `json:"version,omitempty" yaml:"version,omitempty"`
	//NodeSelector for boom deployment
	NodeSelector map[string]string `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	//Tolerations to run boom on nodes
	Tolerations k8s.Tolerations `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`
	//Resource requirements
	Resources *k8s.Resources `json:"resources,omitempty" yaml:"resources,omitempty"`
	//Use this registry to pull the BOOM image from
	//@default: ghcr.io
	CustomImageRegistry string `json:"customImageRegistry,omitempty" yaml:"customImageRegistry,omitempty"`
	//Flag if boom should reconcile itself
	SelfReconciling bool `json:"selfReconciling" yaml:"selfReconciling"`
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
