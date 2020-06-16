package v1beta2

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/monitoring"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/reconciling"
)

type Metadata struct {
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	Namespace string `json:"namespace,omitempty" yaml:"namespace,omitempty"`
}

type ToolsetSpec struct {
	BoomVersion            string                   `json:"boomVersion,omitempty" yaml:"boomVersion,omitempty"`
	ForceApply             bool                     `json:"forceApply,omitempty" yaml:"forceApply,omitempty"`
	CurrentStateFolder     string                   `json:"currentStatePath,omitempty" yaml:"currentStatePath,omitempty"`
	PreApply               *Apply                   `json:"preApply,omitempty" yaml:"preApply,omitempty"`
	PostApply              *Apply                   `json:"postApply,omitempty" yaml:"postApply,omitempty"`
	MetricCollection       *MetricCollection        `json:"metricCollection,omitempty" yaml:"metricCollection,omitempty"`
	LogCollection          *LogCollection           `json:"logCollection,omitempty" yaml:"logCollection,omitempty"`
	NodeMetricsExporter    *NodeMetricsExporter     `json:"nodeMetricsExporter,omitempty" yaml:"nodeMetricsExporter,omitempty"`
	SystemdMetricsExporter *SystemdMetricsExporter  `json:"systemdMetricsExporter,omitempty" yaml:"systemdMetricsExporter,omitempty"`
	Monitoring             *monitoring.Monitoring   `json:"monitoring,omitempty" yaml:"monitoring,omitempty"`
	APIGateway             *APIGateway              `json:"apiGateway,omitempty" yaml:"apiGateway,omitempty"`
	KubeMetricsExporter    *KubeMetricsExporter     `json:"kubeMetricsExporter,omitempty" yaml:"kubeMetricsExporter,omitempty"`
	Reconciling            *reconciling.Reconciling `json:"reconciling,omitempty" yaml:"reconciling,omitempty"`
	MetricsPersisting      *MetricsPersisting       `json:"metricsPersisting,omitempty" yaml:"metricsPersisting,omitempty"`
	LogsPersisting         *LogsPersisting          `json:"logsPersisting,omitempty" yaml:"logsPersisting,omitempty"`
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
