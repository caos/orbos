package metriccollection

import (
	toolsetslatest "github.com/caos/orbos/v5/internal/operator/boom/api/latest"
	"github.com/caos/orbos/v5/internal/operator/boom/application/applications/metriccollection/helm"
	"github.com/caos/orbos/v5/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/v5/internal/utils/helper"
	"github.com/caos/orbos/v5/mntr"
)

func (p *PrometheusOperator) SpecToHelmValues(monitor mntr.Monitor, toolset *toolsetslatest.ToolsetSpec) interface{} {
	// spec := toolset.PrometheusNodeExporter
	imageTags := p.GetImageTags()
	image := "quay.io/coreos/prometheus-operator"

	if toolset != nil && toolset.MetricCollection != nil {
		helper.OverwriteExistingValues(imageTags, map[string]string{
			image: toolset.MetricCollection.OverwriteVersion,
		})
		helper.OverwriteExistingKey(imageTags, &image, toolset.MetricCollection.OverwriteImage)
	}
	values := helm.DefaultValues(imageTags, image)

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
