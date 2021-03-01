package loki

import (
	toolsetslatest "github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/internal/operator/boom/application/applications/loki/helm"
	"github.com/caos/orbos/internal/operator/boom/application/applications/loki/info"
	"github.com/caos/orbos/internal/utils/helper"
	"github.com/caos/orbos/mntr"

	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
)

func (l *Loki) SpecToHelmValues(monitor mntr.Monitor, toolset *toolsetslatest.ToolsetSpec) interface{} {

	imageTags := l.GetImageTags()
	helper.OverwriteExistingValues(imageTags, map[string]string{
		"grafana/loki": toolset.LogsPersisting.OverwriteVersion,
	})
	values := helm.DefaultValues(imageTags)

	values.FullNameOverride = info.GetName().String()

	spec := toolset.LogsPersisting
	if spec == nil {
		return values
	}

	if spec.Storage != nil {
		values.Persistence.Enabled = true
		values.Persistence.Size = spec.Storage.Size
		values.Persistence.StorageClassName = spec.Storage.StorageClass
		if spec.Storage.AccessModes != nil {
			values.Persistence.AccessModes = spec.Storage.AccessModes
		}
	}

	if spec.NodeSelector != nil {
		for k, v := range spec.NodeSelector {
			values.NodeSelector[k] = v
		}
	}

	if spec.Tolerations != nil {
		for _, tol := range spec.Tolerations {
			values.Tolerations = append(values.Tolerations, tol)
		}
	}

	if spec.Resources != nil {
		values.Resources = spec.Resources
	}

	return values
}

func (l *Loki) GetChartInfo() *chart.Chart {
	return helm.GetChartInfo()
}

func (l *Loki) GetImageTags() map[string]string {
	return helm.GetImageTags()
}
