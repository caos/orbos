package nodemetricsexporter

import (
	toolsetslatest "github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/internal/operator/boom/application/applications/nodemetricsexporter/info"
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/mntr"
)

type PrometheusNodeExporter struct {
	monitor mntr.Monitor
}

func New(monitor mntr.Monitor) *PrometheusNodeExporter {
	pne := &PrometheusNodeExporter{
		monitor: monitor,
	}

	return pne
}

func (pne *PrometheusNodeExporter) GetName() name.Application {
	return info.GetName()
}

func (pne *PrometheusNodeExporter) Deploy(toolsetCRDSpec *toolsetslatest.ToolsetSpec) bool {
	return toolsetCRDSpec.NodeMetricsExporter != nil && toolsetCRDSpec.NodeMetricsExporter.Deploy
}

func (pne *PrometheusNodeExporter) GetNamespace() string {
	return info.GetNamespace()
}
