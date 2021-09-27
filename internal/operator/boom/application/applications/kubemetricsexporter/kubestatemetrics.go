package kubemetricsexporter

import (
	toolsetslatest "github.com/caos/orbos/v5/internal/operator/boom/api/latest"
	"github.com/caos/orbos/v5/internal/operator/boom/application/applications/kubemetricsexporter/info"
	"github.com/caos/orbos/v5/internal/operator/boom/name"
	"github.com/caos/orbos/v5/mntr"
)

type KubeStateMetrics struct {
	monitor mntr.Monitor
}

func New(monitor mntr.Monitor) *KubeStateMetrics {
	lo := &KubeStateMetrics{
		monitor: monitor,
	}

	return lo
}

func (k *KubeStateMetrics) GetName() name.Application {
	return info.GetName()
}

func (k *KubeStateMetrics) Deploy(toolsetCRDSpec *toolsetslatest.ToolsetSpec) bool {
	return toolsetCRDSpec.KubeMetricsExporter != nil && toolsetCRDSpec.KubeMetricsExporter.Deploy
}

func (k *KubeStateMetrics) GetNamespace() string {
	return info.GetNamespace()
}
