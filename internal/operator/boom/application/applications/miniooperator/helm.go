package miniooperator

import (
	toolsetslatest "github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/internal/operator/boom/application/applications/miniooperator/helm"
	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/internal/utils/helper"
	"github.com/caos/orbos/mntr"
)

func (m *MinioOperator) SpecToHelmValues(monitor mntr.Monitor, toolset *toolsetslatest.ToolsetSpec) interface{} {
	imageTags := m.GetImageTags()
	image := "minio/operator"

	if toolset != nil && toolset.S3StorageOperator != nil {
		helper.OverwriteExistingValues(imageTags, map[string]string{
			image: toolset.S3StorageOperator.OverwriteVersion,
		})
		helper.OverwriteExistingKey(imageTags, &image, toolset.S3StorageOperator.OverwriteImage)
	}
	values := helm.DefaultValues(imageTags, image)

	spec := toolset.S3StorageOperator
	if spec == nil {
		return values
	}

	if spec.ClusterDomain != "" {
		values.Operator.ClusterDomain = spec.ClusterDomain
	}

	if spec.Resources != nil {
		values.Operator.Resources = spec.Resources
	}

	return values
}
func (m *MinioOperator) GetChartInfo() *chart.Chart {
	return helm.GetChartInfo()
}

func (m *MinioOperator) GetImageTags() map[string]string {
	return helm.GetImageTags()
}
