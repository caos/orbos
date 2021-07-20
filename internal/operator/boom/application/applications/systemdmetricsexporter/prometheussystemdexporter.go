package systemdmetricsexporter

import (
	toolsetslatest "github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/internal/operator/boom/application/applications/systemdmetricsexporter/info"
	"github.com/caos/orbos/internal/operator/boom/name"
)

type prometheusSystemdExporter struct{}

func New() *prometheusSystemdExporter {
	return &prometheusSystemdExporter{}
}

func (*prometheusSystemdExporter) GetName() name.Application {
	return info.GetName()
}

func (*prometheusSystemdExporter) Deploy(toolsetCRDSpec *toolsetslatest.ToolsetSpec) bool {
	return toolsetCRDSpec.SystemdMetricsExporter != nil && toolsetCRDSpec.SystemdMetricsExporter.Deploy
}

func (*prometheusSystemdExporter) GetNamespace() string {
	return info.GetNamespace()
}
