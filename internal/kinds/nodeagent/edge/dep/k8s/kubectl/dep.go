package kubectl

import (
	"github.com/caos/infrop/internal/core/operator"
	"github.com/caos/infrop/internal/kinds/nodeagent/adapter"
	"github.com/caos/infrop/internal/kinds/nodeagent/edge/dep"
	"github.com/caos/infrop/internal/kinds/nodeagent/edge/dep/k8s"
	"github.com/caos/infrop/internal/kinds/nodeagent/edge/dep/middleware"
)

type Installer interface {
	isKubectl()
	adapter.Installer
}

type kubectlDep struct {
	common *k8s.Common
}

func New(os dep.OperatingSystem, manager *dep.PackageManager) Installer {
	return &kubectlDep{k8s.New(os, manager, "kubectl")}
}

func (kubectlDep) isKubectl() {}

func (kubectlDep) Is(other adapter.Installer) bool {
	_, ok := middleware.Unwrap(other).(Installer)
	return ok
}

func (k kubectlDep) String() string { return "Kubectl" }

func (*kubectlDep) Equals(other adapter.Installer) bool {
	_, ok := other.(*kubectlDep)
	return ok
}

func (k *kubectlDep) Current() (operator.Package, error) {
	return k.common.Current()
}

func (k *kubectlDep) Ensure(remove operator.Package, install operator.Package) (bool, error) {
	return false, k.common.Ensure(remove, install)
}
