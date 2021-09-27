package metriccollection

import (
	toolsetslatest "github.com/caos/orbos/v5/internal/operator/boom/api/latest"
	"github.com/caos/orbos/v5/internal/operator/boom/application/applications/metriccollection/info"
	"github.com/caos/orbos/v5/internal/operator/boom/name"
	"github.com/caos/orbos/v5/mntr"
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

func (po *PrometheusOperator) Deploy(toolsetCRDSpec *toolsetslatest.ToolsetSpec) bool {
	return toolsetCRDSpec.MetricCollection != nil && toolsetCRDSpec.MetricCollection.Deploy
}

func (po *PrometheusOperator) GetNamespace() string {
	return info.GetNamespace()
}
