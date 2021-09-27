package apigateway

import (
	toolsetslatest "github.com/caos/orbos/v5/internal/operator/boom/api/latest"
	"github.com/caos/orbos/v5/internal/operator/boom/application/applications/apigateway/info"
	"github.com/caos/orbos/v5/internal/operator/boom/name"
	"github.com/caos/orbos/v5/mntr"
)

type Ambassador struct {
	monitor mntr.Monitor
}

func New(monitor mntr.Monitor) *Ambassador {
	return &Ambassador{
		monitor: monitor,
	}
}

func (a *Ambassador) Deploy(toolsetCRDSpec *toolsetslatest.ToolsetSpec) bool {
	return toolsetCRDSpec.APIGateway != nil && toolsetCRDSpec.APIGateway.Deploy
}

func (a *Ambassador) GetName() name.Application {
	return info.GetName()
}

func (a *Ambassador) GetNamespace() string {
	return info.GetNamespace()
}
