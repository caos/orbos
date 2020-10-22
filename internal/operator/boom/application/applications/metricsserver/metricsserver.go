package metricsserver

import (
	toolsetsv1beta2 "github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	"github.com/caos/orbos/internal/operator/boom/application/applications/metricsserver/info"
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/mntr"
)

type MetricsServer struct {
	monitor mntr.Monitor
}

func New(monitor mntr.Monitor) *MetricsServer {
	po := &MetricsServer{
		monitor: monitor,
	}

	return po
}

func (po *MetricsServer) GetName() name.Application {
	return info.GetName()
}

func (po *MetricsServer) Deploy(toolsetCRDSpec *toolsetsv1beta2.ToolsetSpec) bool {
	return toolsetCRDSpec.MetricsServer != nil && toolsetCRDSpec.MetricsServer.Deploy
}

func (po *MetricsServer) GetNamespace() string {
	return info.GetNamespace()
}
