package kubestatemetrics

import (
	toolsetsv1beta1 "github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	"github.com/caos/orbos/internal/operator/boom/application/applications/kubestatemetrics/info"
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/mntr"
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

func (k *KubeStateMetrics) Deploy(toolsetCRDSpec *toolsetsv1beta1.ToolsetSpec) bool {
	return toolsetCRDSpec.KubeStateMetrics.Deploy
}

func (k *KubeStateMetrics) GetNamespace() string {
	return info.GetNamespace()
}
