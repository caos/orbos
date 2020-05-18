package prometheussystemdexporter

import (
	toolsetsv1beta1 "github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheussystemdexporter/info"
	"github.com/caos/orbos/internal/operator/boom/name"
)

type prometheusSystemdExporter struct{}

func New() *prometheusSystemdExporter {
	return &prometheusSystemdExporter{}
}

func (*prometheusSystemdExporter) GetName() name.Application {
	return info.GetName()
}

func (*prometheusSystemdExporter) Deploy(toolsetCRDSpec *toolsetsv1beta1.ToolsetSpec) bool {
	return toolsetCRDSpec.PrometheusSystemdExporter == nil || toolsetCRDSpec.PrometheusSystemdExporter.Deploy
}

func (*prometheusSystemdExporter) GetNamespace() string {
	return info.GetNamespace()
}
