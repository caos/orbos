package prometheusnodeexporter

import (
	toolsetsv1beta2 "github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheusnodeexporter/info"
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

func (pne *PrometheusNodeExporter) Deploy(toolsetCRDSpec *toolsetsv1beta2.ToolsetSpec) bool {
	return toolsetCRDSpec.NodeMetricsExporter.Deploy
}

func (pne *PrometheusNodeExporter) GetNamespace() string {
	return info.GetNamespace()
}
