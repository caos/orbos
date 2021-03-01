package prometheusnodeexporter

import (
	toolsetslatest "github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheusnodeexporter/helm"
	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/internal/utils/helper"
	"github.com/caos/orbos/mntr"
)

func (p *PrometheusNodeExporter) SpecToHelmValues(monitor mntr.Monitor, toolset *toolsetslatest.ToolsetSpec) interface{} {
	// spec := toolset.PrometheusNodeExporter
	imageTags := p.GetImageTags()
	if toolset != nil && toolset.NodeMetricsExporter != nil {
		helper.OverwriteExistingValues(imageTags, map[string]string{
			"quay.io/prometheus/node-exporter": toolset.NodeMetricsExporter.OverwriteVersion,
		})
	}
	values := helm.DefaultValues(imageTags)

	// if spec.ReplicaCount != 0 {
	// 	values.ReplicaCount = spec.ReplicaCount
	// }

	spec := toolset.NodeMetricsExporter

	if spec == nil {
		return values
	}

	if spec.Resources != nil {
		values.Resources = spec.Resources
	}

	return values
}

func (p *PrometheusNodeExporter) GetChartInfo() *chart.Chart {
	return helm.GetChartInfo()
}

func (c *PrometheusNodeExporter) GetImageTags() map[string]string {
	return helm.GetImageTags()
}
