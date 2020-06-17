package loggingoperator

import (
	toolsetsv1beta1 "github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	"github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/info"
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/mntr"
)

type LoggingOperator struct {
	monitor mntr.Monitor
}

func New(monitor mntr.Monitor) *LoggingOperator {
	lo := &LoggingOperator{
		monitor: monitor,
	}

	return lo
}
func (l *LoggingOperator) GetName() name.Application {
	return info.GetName()
}

func (lo *LoggingOperator) Deploy(toolsetCRDSpec *toolsetsv1beta1.ToolsetSpec) bool {
	return toolsetCRDSpec.LoggingOperator != nil && toolsetCRDSpec.LoggingOperator.Deploy
}

func (l *LoggingOperator) GetNamespace() string {
	return info.GetNamespace()
}
