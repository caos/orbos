package metricsserver

import (
	toolsetslatest "github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/internal/operator/boom/application/applications/metricsserver/helm"
	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/internal/utils/helper"
	"github.com/caos/orbos/mntr"
)

func (m *MetricsServer) HelmMutate(monitor mntr.Monitor, toolsetCRDSpec *toolsetslatest.ToolsetSpec, resultFilePath string) error {

	if err := helper.DeleteFirstResourceFromYaml(resultFilePath, "v1", "Pod", "metrics-server-test"); err != nil {
		return err
	}

	return nil
}

func (m *MetricsServer) SpecToHelmValues(monitor mntr.Monitor, toolset *toolsetslatest.ToolsetSpec) interface{} {
	imageTags := m.GetImageTags()
	if toolset != nil && toolset.MetricsServer != nil {
		helper.OverwriteExistingValues(imageTags, map[string]string{
			"k8s.gcr.io/metrics-server-amd64": toolset.MetricsServer.OverwriteVersion,
		})
	}
	values := helm.DefaultValues(imageTags)

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
