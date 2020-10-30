package metricsserver

import (
	toolsetsv1beta2 "github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	"github.com/caos/orbos/internal/operator/boom/application/applications/metricsserver/helm"
	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/internal/utils/helper"
	"github.com/caos/orbos/mntr"
)

func (m *MetricsServer) HelmMutate(monitor mntr.Monitor, toolsetCRDSpec *toolsetsv1beta2.ToolsetSpec, resultFilePath string) error {

	if err := helper.DeleteFirstResourceFromYaml(resultFilePath, "v1", "Pod", "metrics-server-test"); err != nil {
		return err
	}

	return nil
}

func (m *MetricsServer) SpecToHelmValues(monitor mntr.Monitor, toolset *toolsetsv1beta2.ToolsetSpec) interface{} {
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
