package nginx

import (
	"io/ioutil"
	"os"

	"github.com/caos/infrop/internal/core/operator"
	"github.com/caos/infrop/internal/kinds/nodeagent/adapter"
	"github.com/caos/infrop/internal/kinds/nodeagent/edge/dep"
	"github.com/caos/infrop/internal/kinds/nodeagent/edge/dep/middleware"
)

type Installer interface {
	isNgninx()
	adapter.Installer
}
type nginxDep struct {
	manager *dep.PackageManager
	systemd *dep.SystemD
}

func New(manager *dep.PackageManager, systemd *dep.SystemD) Installer {
	return &nginxDep{manager, systemd}
}

func (nginxDep) isNgninx() {}

func (nginxDep) Is(other adapter.Installer) bool {
	_, ok := middleware.Unwrap(other).(Installer)
	return ok
}

func (nginxDep) String() string { return "NGINX" }

func (*nginxDep) Equals(other adapter.Installer) bool {
	_, ok := other.(*nginxDep)
	return ok
}

func (s *nginxDep) Current() (pkg operator.Package, err error) {
	config, err := ioutil.ReadFile("/etc/nginx/nginx.conf")
	if os.IsNotExist(err) {
		return pkg, nil
	}

	pkg.Config = map[string]string{
		"nginx.conf": string(config),
	}
	return pkg, err
}

func (s *nginxDep) Ensure(remove operator.Package, ensure operator.Package) (bool, error) {

	ensureCfg, ok := ensure.Config["nginx.conf"]
	if !ok {
		s.systemd.Disable("nginx")
		os.Remove("/etc/nginx/nginx.conf")
		return false, nil
	}

	if _, ok := remove.Config["nginx.conf"]; !ok {
		if err := s.manager.Install(&dep.Software{
			Package: "nginx",
			Version: ensure.Version,
		}); err != nil {
			return false, err
		}

		if err := os.MkdirAll("/etc/nginx", 0700); err != nil {
			return false, err
		}

		if err := ioutil.WriteFile("/proc/sys/net/ipv4/ip_forward", []byte("1"), 0600); err != nil {
			return false, err
		}
	}

	if err := ioutil.WriteFile("/etc/nginx/nginx.conf", []byte(ensureCfg), 0600); err != nil {
		return false, err
	}

	return false, s.systemd.Enable("nginx")
}
