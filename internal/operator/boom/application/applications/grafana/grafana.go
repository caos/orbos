package grafana

import (
	"reflect"

	toolsetslatest "github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/internal/operator/boom/api/latest/monitoring"

	"github.com/caos/orbos/internal/operator/boom/application/applications/grafana/info"
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/mntr"
)

type Grafana struct {
	monitor mntr.Monitor
	spec    *monitoring.Monitoring
}

func New(monitor mntr.Monitor) *Grafana {
	return &Grafana{
		monitor: monitor,
	}
}
func (g *Grafana) GetName() name.Application {
	return info.GetName()
}

func (g *Grafana) Deploy(toolsetCRDSpec *toolsetslatest.ToolsetSpec) bool {
	// workaround as grafana always deploys new pods even when the spec of the deployment is not changed
	// due to the fact that kubernetes has an internal mapping from extensions/v1beta1 to apps/v1 in old k8s versions
	if g.Changed(toolsetCRDSpec) {
		return toolsetCRDSpec.Monitoring != nil && toolsetCRDSpec.Monitoring.Deploy
	}
	return false
}

func (g *Grafana) Changed(toolsetCRDSpec *toolsetslatest.ToolsetSpec) bool {
	if g.spec == nil {
		return true
	}
	return !reflect.DeepEqual(toolsetCRDSpec.Monitoring, g.spec)
}

func (g *Grafana) SetAppliedSpec(toolsetCRDSpec *toolsetslatest.ToolsetSpec) {
	g.spec = toolsetCRDSpec.Monitoring
}

func (g *Grafana) GetNamespace() string {
	return info.GetNamespace()
}
