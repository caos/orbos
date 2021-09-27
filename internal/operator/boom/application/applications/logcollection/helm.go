package logcollection

import (
	toolsetslatest "github.com/caos/orbos/v5/internal/operator/boom/api/latest"
	"github.com/caos/orbos/v5/internal/operator/boom/application/applications/logcollection/helm"
	"github.com/caos/orbos/v5/internal/operator/boom/application/applications/logspersisting/logs"
	"github.com/caos/orbos/v5/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/v5/internal/utils/helper"
	"github.com/caos/orbos/v5/mntr"
)

func (l *LoggingOperator) HelmPreApplySteps(monitor mntr.Monitor, toolsetCRDSpec *toolsetslatest.ToolsetSpec) ([]interface{}, error) {
	return logs.GetAllResources(toolsetCRDSpec), nil
}

func (l *LoggingOperator) SpecToHelmValues(monitor mntr.Monitor, toolset *toolsetslatest.ToolsetSpec) interface{} {
	// spec := toolset.LoggingOperator
	imageTags := l.GetImageTags()
	image := "banzaicloud/logging-operator"

	if toolset != nil && toolset.LogCollection != nil {
		helper.OverwriteExistingValues(imageTags, map[string]string{
			image: toolset.LogCollection.OverwriteVersion,
		})
		helper.OverwriteExistingKey(imageTags, &image, toolset.LogCollection.OverwriteImage)
	}
	values := helm.DefaultValues(imageTags, image)

	// if spec.ReplicaCount != 0 {
	// 	values.ReplicaCount = spec.ReplicaCount
	// }

	spec := toolset.LogCollection
	if spec == nil || spec.Operator == nil {
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
