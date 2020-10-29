package loggingoperator

import (
	toolsetslatest "github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/helm"
	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/mntr"
)

func (l *LoggingOperator) SpecToHelmValues(monitor mntr.Monitor, toolset *toolsetslatest.ToolsetSpec) interface{} {
	// spec := toolset.LoggingOperator
	values := helm.DefaultValues(l.GetImageTags())

	// if spec.ReplicaCount != 0 {
	// 	values.ReplicaCount = spec.ReplicaCount
	// }

	spec := toolset.LogCollection
	if spec == nil {
		return values
	}

	if spec.Operator.NodeSelector != nil {
		for k, v := range spec.Operator.NodeSelector {
			values.NodeSelector[k] = v
		}
	}

	if spec.Operator.Tolerations != nil {
		for _, tol := range spec.Operator.Tolerations {
			values.Tolerations = append(values.Tolerations, tol)
		}
	}

	if spec.Operator.Resources != nil {
		values.Resources = spec.Operator.Resources
	}

	return values
}

func (l *LoggingOperator) GetChartInfo() *chart.Chart {
	return helm.GetChartInfo()
}

func (l *LoggingOperator) GetImageTags() map[string]string {
	return helm.GetImageTags()
}
