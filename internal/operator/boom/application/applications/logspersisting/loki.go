package logspersisting

import (
	toolsetslatest "github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/internal/operator/boom/application/applications/logspersisting/info"
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/mntr"
)

type Loki struct {
	monitor mntr.Monitor
}

func New(monitor mntr.Monitor) *Loki {
	lo := &Loki{
		monitor: monitor,
	}
	return lo
}

func (l *Loki) GetName() name.Application {
	return info.GetName()
}

func (lo *Loki) Deploy(toolsetCRDSpec *toolsetslatest.ToolsetSpec) bool {
	return toolsetCRDSpec.LogsPersisting != nil && toolsetCRDSpec.LogsPersisting.Deploy
}

func (l *Loki) GetNamespace() string {
	return info.GetNamespace()
}
