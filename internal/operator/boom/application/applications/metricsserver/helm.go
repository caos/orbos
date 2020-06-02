package metricsserver

import (
	toolsetsv1beta1 "github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	"github.com/caos/orbos/internal/operator/boom/application/applications/metricsserver/helm"
	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/mntr"
)

func (m *MetricsServer) SpecToHelmValues(monitor mntr.Monitor, toolset *toolsetsv1beta1.ToolsetSpec) interface{} {
	values := helm.DefaultValues(m.GetImageTags())

	// if spec.ReplicaCount != 0 {
	// 	values.ReplicaCount = toolset.MetricsServer.ReplicaCount
	// }

	return values
}

func (m *MetricsServer) GetChartInfo() *chart.Chart {
	return helm.GetChartInfo()
}

func (m *MetricsServer) GetImageTags() map[string]string {
	return helm.GetImageTags()
}
