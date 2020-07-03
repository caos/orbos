package loggingoperator

import (
	toolsetsv1beta2 "github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	"github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/helm"
	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/mntr"
)

func (l *LoggingOperator) SpecToHelmValues(monitor mntr.Monitor, toolset *toolsetsv1beta2.ToolsetSpec) interface{} {
	// spec := toolset.LoggingOperator
	values := helm.DefaultValues(l.GetImageTags())

	// if spec.ReplicaCount != 0 {
	// 	values.ReplicaCount = spec.ReplicaCount
	// }

	return values
}

func (l *LoggingOperator) GetChartInfo() *chart.Chart {
	return helm.GetChartInfo()
}

func (l *LoggingOperator) GetImageTags() map[string]string {
	return helm.GetImageTags()
}
