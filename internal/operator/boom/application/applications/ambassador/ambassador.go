package ambassador

import (
	toolsetsv1beta2 "github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	"github.com/caos/orbos/internal/operator/boom/application/applications/ambassador/info"
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/mntr"
)

type Ambassador struct {
	monitor mntr.Monitor
}

func New(monitor mntr.Monitor) *Ambassador {
	return &Ambassador{
		monitor: monitor,
	}
}

func (a *Ambassador) Deploy(toolsetCRDSpec *toolsetsv1beta2.ToolsetSpec) bool {
	return toolsetCRDSpec.APIGateway.Deploy
}

func (a *Ambassador) GetName() name.Application {
	return info.GetName()
}

func (a *Ambassador) GetNamespace() string {
	return info.GetNamespace()
}
