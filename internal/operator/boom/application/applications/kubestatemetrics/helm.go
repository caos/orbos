package kubestatemetrics

import (
	toolsetsv1beta2 "github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	"github.com/caos/orbos/internal/operator/boom/application/applications/kubestatemetrics/helm"
	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/mntr"
)

func (k *KubeStateMetrics) SpecToHelmValues(monitor mntr.Monitor, toolset *toolsetsv1beta2.ToolsetSpec) interface{} {
	values := helm.DefaultValues(k.GetImageTags())

	spec := toolset.KubeMetricsExporter

	if spec == nil {
		return values
	}

	if spec.ReplicaCount != 0 {
		values.Replicas = spec.ReplicaCount
	}

	if spec.NodeSelector != nil {
		for k, v := range spec.NodeSelector {
			values.NodeSelector[k] = v
		}
	}

	if spec.Tolerations != nil {
		for _, tol := range spec.Tolerations {
			values.Tolerations = append(values.Tolerations, tol)
		}
	}

	if spec.Resources != nil {
		values.Resources = spec.Resources
	}

	return values
}

func (k *KubeStateMetrics) GetChartInfo() *chart.Chart {
	return helm.GetChartInfo()
}

func (k *KubeStateMetrics) GetImageTags() map[string]string {
	return helm.GetImageTags()
}
