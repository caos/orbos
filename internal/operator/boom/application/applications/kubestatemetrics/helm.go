package kubestatemetrics

import (
	toolsetsv1beta2 "github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	"github.com/caos/orbos/internal/operator/boom/application/applications/kubestatemetrics/helm"
	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/mntr"
)

func (k *KubeStateMetrics) SpecToHelmValues(monitor mntr.Monitor, toolset *toolsetsv1beta2.ToolsetSpec) interface{} {
	values := helm.DefaultValues(k.GetImageTags())

	if toolset.KubeMetricsExporter == nil {
		return values
	}

	if toolset.KubeMetricsExporter.ReplicaCount != 0 {
		values.Replicas = toolset.KubeMetricsExporter.ReplicaCount
	}

	if toolset.KubeMetricsExporter.NodeSelector != nil {
		for k, v := range toolset.KubeMetricsExporter.NodeSelector {
			values.NodeSelector[k] = v
		}
	}

	if toolset.KubeMetricsExporter.Tolerations != nil {
		for _, tol := range toolset.KubeMetricsExporter.Tolerations {
			values.Tolerations = append(values.Tolerations, tol.ToKubeToleration())
		}
	}

	return values
}

func (k *KubeStateMetrics) GetChartInfo() *chart.Chart {
	return helm.GetChartInfo()
}

func (k *KubeStateMetrics) GetImageTags() map[string]string {
	return helm.GetImageTags()
}
