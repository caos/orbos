package nginx

import (
	"io/ioutil"
	"os"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/nodeagent"
	"github.com/caos/orbiter/internal/operator/nodeagent/dep"
	"github.com/caos/orbiter/internal/operator/nodeagent/dep/middleware"
	"github.com/caos/orbiter/mntr"
)

type Installer interface {
	isNgninx()
	nodeagent.Installer
}
type nginxDep struct {
	manager *dep.PackageManager
	systemd *dep.SystemD
	monitor mntr.Monitor
}

func New(monitor mntr.Monitor, manager *dep.PackageManager, systemd *dep.SystemD) Installer {
	return &nginxDep{manager, systemd, monitor}
}

func (nginxDep) isNgninx() {}

func (nginxDep) Is(other nodeagent.Installer) bool {
	_, ok := middleware.Unwrap(other).(Installer)
	return ok
}

func (nginxDep) String() string { return "NGINX" }

func (*nginxDep) Equals(other nodeagent.Installer) bool {
	_, ok := other.(*nginxDep)
	return ok
}

func (s *nginxDep) Current() (pkg common.Package, err error) {
	config, err := ioutil.ReadFile("/etc/nginx/nginx.conf")
	if os.IsNotExist(err) {
		return pkg, nil
	}

	pkg.Config = map[string]string{
		"nginx.conf": string(config),
	}

	return pkg, nil
}

func (s *nginxDep) Ensure(remove common.Package, ensure common.Package) error {

	ensureCfg, ok := ensure.Config["nginx.conf"]
	if !ok {
		s.systemd.Disable("nginx")
		os.Remove("/etc/nginx/nginx.conf")
		return nil
	}

	if _, ok := remove.Config["nginx.conf"]; !ok {

		if err := ioutil.WriteFile("/etc/yum.repos.d/nginx.repo", []byte(`[nginx-stable]
name=nginx stable repo
baseurl=http://nginx.org/packages/centos/$releasever/$basearch/
gpgcheck=1
enabled=1
gpgkey=https://nginx.org/keys/nginx_signing.key
module_hotfixes=true`), 0600); err != nil {
			return err
		}

		if err := s.manager.Install(&dep.Software{
			Package: "nginx",
			Version: ensure.Version,
		}); err != nil {
			return err
		}

		if err := os.MkdirAll("/etc/nginx", 0700); err != nil {
			return err
		}
	}

	if err := ioutil.WriteFile("/etc/nginx/nginx.conf", []byte(ensureCfg), 0600); err != nil {
		return err
	}

	if err := s.systemd.Enable("nginx"); err != nil {
		return err
	}

	return s.systemd.Start("nginx")
}
