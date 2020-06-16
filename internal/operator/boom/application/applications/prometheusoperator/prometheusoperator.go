package prometheusoperator

import (
	toolsetsv1beta2 "github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheusoperator/info"
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/mntr"
)

type PrometheusOperator struct {
	monitor mntr.Monitor
}

func New(monitor mntr.Monitor) *PrometheusOperator {
	po := &PrometheusOperator{
		monitor: monitor,
	}

	return po
}

func (po *PrometheusOperator) GetName() name.Application {
	return info.GetName()
}

func (po *PrometheusOperator) Deploy(toolsetCRDSpec *toolsetsv1beta2.ToolsetSpec) bool {
	return toolsetCRDSpec.MetricCollection.Deploy
}

func (po *PrometheusOperator) GetNamespace() string {
	return info.GetNamespace()
}
