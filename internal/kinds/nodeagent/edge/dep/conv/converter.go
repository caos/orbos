package conv

import (
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/kinds/nodeagent/adapter"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/dep"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/dep/cri"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/dep/hostname"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/dep/k8s/kubeadm"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/dep/k8s/kubectl"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/dep/k8s/kubelet"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/dep/keepalived"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/dep/middleware"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/dep/nginx"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/dep/swap"
	"github.com/caos/orbiter/logging"
)

type Converter interface {
	Init() (adapter.Callback, error)
	adapter.Converter
}

type dependencies struct {
	logger logging.Logger
	os     dep.OperatingSystemMajor
	pm     *dep.PackageManager
	sysd   *dep.SystemD
	cipher string
}

func New(logger logging.Logger, os dep.OperatingSystemMajor, cipher string) Converter {
	return &dependencies{logger, os, nil, nil, cipher}
}

func (d *dependencies) Init() (adapter.Callback, error) {

	d.pm = dep.NewPackageManager(d.logger, d.os.OperatingSystem)
	if err := d.pm.Init(); err != nil {
		return d.pm.RefreshInstalled, err
	}

	d.sysd = dep.NewSystemD(d.logger)
	return d.pm.RefreshInstalled, nil
}

func (d *dependencies) ToDependencies(sw operator.Software) []*adapter.Dependency {

	dependencies := []*adapter.Dependency{
		&adapter.Dependency{
			Desired:   sw.Hostname,
			Installer: hostname.New(),
		},
		&adapter.Dependency{
			Desired:   sw.Swap,
			Installer: swap.New("/etc/fstab"),
		},
		&adapter.Dependency{
			Desired:   sw.KeepaliveD,
			Installer: keepalived.New(d.logger, d.pm, d.sysd, d.cipher),
		},
		&adapter.Dependency{
			Desired:   sw.Nginx,
			Installer: nginx.New(d.logger, d.pm, d.sysd),
		},
		&adapter.Dependency{
			Desired:   sw.Containerruntime,
			Installer: cri.New(d.logger, d.os, d.pm),
		},
		&adapter.Dependency{
			Desired:   sw.Kubelet,
			Installer: kubelet.New(d.logger, d.os.OperatingSystem, d.pm, d.sysd),
		},
		&adapter.Dependency{
			Desired:   sw.Kubectl,
			Installer: kubectl.New(d.os.OperatingSystem, d.pm),
		},
		&adapter.Dependency{
			Desired:   sw.Kubeadm,
			Installer: kubeadm.New(d.os.OperatingSystem, d.pm),
		},
	}

	for key, dependency := range dependencies {
		dependency.Installer = middleware.AddLogging(d.logger, dependency.Installer)
		dependencies[key] = dependency
	}

	return dependencies
}

func (d *dependencies) ToSoftware(dependencies []*adapter.Dependency) (sw operator.Software) {

	for _, dependency := range dependencies {
		switch i := middleware.Unwrap(dependency.Installer).(type) {
		case hostname.Installer:
			sw.Hostname = dependency.Current
		case swap.Installer:
			sw.Swap = dependency.Current
		case kubelet.Installer:
			sw.Kubelet = dependency.Current
		case kubeadm.Installer:
			sw.Kubeadm = dependency.Current
		case kubectl.Installer:
			sw.Kubectl = dependency.Current
		case cri.Installer:
			sw.Containerruntime = dependency.Current
		case keepalived.Installer:
			sw.KeepaliveD = dependency.Current
		case nginx.Installer:
			sw.Nginx = dependency.Current
		default:
			panic(errors.Errorf("No installer type for dependency %s found", i))
		}
	}

	return sw
}
