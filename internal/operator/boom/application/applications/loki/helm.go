package loki

import (
	toolsetsv1beta2 "github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	"github.com/caos/orbos/internal/operator/boom/application/applications/loki/helm"
	"github.com/caos/orbos/internal/operator/boom/application/applications/loki/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/loki/logs"
	"github.com/caos/orbos/mntr"

	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
)

func (l *Loki) HelmPreApplySteps(monitor mntr.Monitor, toolsetCRDSpec *toolsetsv1beta2.ToolsetSpec) ([]interface{}, error) {
	return logs.GetAllResources(toolsetCRDSpec), nil
}

func (l *Loki) SpecToHelmValues(monitor mntr.Monitor, toolset *toolsetsv1beta2.ToolsetSpec) interface{} {

	values := helm.DefaultValues(l.GetImageTags())

	if toolset.LogsPersisting != nil {
		spec := toolset.LogsPersisting
		if spec.Storage != nil {
			values.Persistence.Enabled = true
			values.Persistence.Size = spec.Storage.Size
			values.Persistence.StorageClassName = spec.Storage.StorageClass
			if spec.Storage.AccessModes != nil {
				values.Persistence.AccessModes = spec.Storage.AccessModes
			}
		}
	}

	if toolset.LogsPersisting.NodeSelector != nil {
		for k, v := range toolset.Monitoring.NodeSelector {
			values.NodeSelector[k] = v
		}
	}

	if toolset.LogsPersisting.Tolerations != nil {
		for _, t := range toolset.LogsPersisting.Tolerations {
			values.Tolerations = append(values.Tolerations, &helm.Toleration{
				Effect:            t.Effect,
				Key:               t.Key,
				Operator:          t.Operator,
				TolerationSeconds: t.TolerationSeconds,
				Value:             t.Value,
			})
		}
	}

	values.FullNameOverride = info.GetName().String()
	return values
}

func (l *Loki) GetChartInfo() *chart.Chart {
	return helm.GetChartInfo()
}

func (l *Loki) GetImageTags() map[string]string {
	return helm.GetImageTags()
}
