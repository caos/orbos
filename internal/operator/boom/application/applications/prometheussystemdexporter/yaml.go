package prometheussystemdexporter

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheussystemdexporter/yaml"
	"github.com/caos/orbos/mntr"
)

// var _ application.YAMLApplication = (*prometheusSystemdExporter)(nil)

func (*prometheusSystemdExporter) GetYaml(_ mntr.Monitor, _ *v1beta1.ToolsetSpec) interface{} {
	return yaml.GetDefault()
}
