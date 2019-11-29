package kubeadm

import (
	"regexp"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/kinds/nodeagent/adapter"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/dep"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/dep/middleware"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/dep/k8s"
)

type Installer interface {
	isKubeadm()
	adapter.Installer
}

type kubeadmDep struct {
	manager    *dep.PackageManager
	os         dep.OperatingSystem
	normalizer *regexp.Regexp
	common     *k8s.Common
}

func New(os dep.OperatingSystem, manager *dep.PackageManager) Installer {
	return &kubeadmDep{manager, os, regexp.MustCompile(`\d+\.\d+\.\d+`), k8s.New(os, manager, "kubeadm")}
}

func (kubeadmDep) isKubeadm() {}

func (kubeadmDep) Is(other adapter.Installer) bool {
	_, ok := middleware.Unwrap(other).(Installer)
	return ok
}

func (k kubeadmDep) String() string { return "Kubeadm" }

func (*kubeadmDep) Equals(other adapter.Installer) bool {
	_, ok := other.(*kubeadmDep)
	return ok
}

func (k *kubeadmDep) Current() (operator.Package, error) {
	return k.common.Current()
}

func (k *kubeadmDep) Ensure(remove operator.Package, install operator.Package) (bool, error) {
	return false, k.common.Ensure(remove, install)
}
