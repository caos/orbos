package prometheusoperator

import (
	toolsetslatest "github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheusoperator/helm"
	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/labels"
)

func (p *PrometheusOperator) SpecToHelmValues(monitor mntr.Monitor, l *labels.API, toolset *toolsetslatest.ToolsetSpec) interface{} {
	// spec := toolset.PrometheusNodeExporter
	values := helm.DefaultValues(p.GetImageTags(), labels.MustForComponent(l, "metricCollection"))

	// if spec.ReplicaCount != 0 {
	// 	values.ReplicaCount = spec.ReplicaCount
	// }

	spec := toolset.MetricCollection
	if spec == nil {
		return values
	}

	if spec.NodeSelector != nil {
		for k, v := range spec.NodeSelector {
			values.PrometheusOperator.NodeSelector[k] = v
		}
	}

	if spec.Tolerations != nil {
		for _, tol := range spec.Tolerations {
			values.PrometheusOperator.Tolerations = append(values.PrometheusOperator.Tolerations, tol)
		}
	}

	if spec.Resources != nil {
		values.PrometheusOperator.Resources = spec.Resources
	}

	return values
}

func (p *PrometheusOperator) GetChartInfo() *chart.Chart {
	return helm.GetChartInfo()
}

func (p *PrometheusOperator) GetImageTags() map[string]string {
	return helm.GetImageTags()
}
