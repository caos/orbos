package miniooperator

import (
	toolsetslatest "github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/internal/operator/boom/application/applications/miniooperator/info"
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/mntr"
)

type MinioOperator struct {
	monitor mntr.Monitor
}

func New(monitor mntr.Monitor) *MinioOperator {
	mo := &MinioOperator{
		monitor: monitor,
	}

	return mo
}

func (po *MinioOperator) GetName() name.Application {
	return info.GetName()
}

func (po *MinioOperator) Deploy(toolsetCRDSpec *toolsetslatest.ToolsetSpec) bool {
	return toolsetCRDSpec.S3StorageOperator != nil && toolsetCRDSpec.S3StorageOperator.Deploy
}

func (po *MinioOperator) GetNamespace() string {
	return info.GetNamespace()
}
