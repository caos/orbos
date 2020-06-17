package argocd

import (
	toolsetsv1beta1 "github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	"github.com/caos/orbos/internal/operator/boom/application/applications/argocd/info"
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/mntr"
)

type Argocd struct {
	monitor mntr.Monitor
}

func New(monitor mntr.Monitor) *Argocd {
	c := &Argocd{
		monitor: monitor,
	}

	return c
}

func (a *Argocd) GetName() name.Application {
	return info.GetName()
}

func (a *Argocd) Deploy(toolsetCRDSpec *toolsetsv1beta1.ToolsetSpec) bool {
	return toolsetCRDSpec.Argocd != nil && toolsetCRDSpec.Argocd.Deploy
}

func (a *Argocd) GetNamespace() string {
	return info.GetNamespace()
}
