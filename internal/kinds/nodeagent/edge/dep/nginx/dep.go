package nginx

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/kinds/nodeagent/adapter"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/dep"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/dep/middleware"
	"github.com/caos/orbiter/logging"
)

type Installer interface {
	isNgninx()
	adapter.Installer
}
type nginxDep struct {
	manager *dep.PackageManager
	systemd *dep.SystemD
	logger  logging.Logger
}

func New(logger logging.Logger, manager *dep.PackageManager, systemd *dep.SystemD) Installer {
	return &nginxDep{manager, systemd, logger}
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

const (
	ipForwardCfg    = "/proc/sys/net/ipv4/ip_forward"
	nonlocalbindCfg = "/proc/sys/net/ipv4/ip_nonlocal_bind"
)

func (s *nginxDep) Current() (pkg operator.Package, err error) {
	config, err := ioutil.ReadFile("/etc/nginx/nginx.conf")
	if os.IsNotExist(err) {
		return pkg, nil
	}

	pkg.Config = map[string]string{
		"nginx.conf": string(config),
	}

	enabled, err := s.currentSysctlConfig("net.ipv4.ip_nonlocal_bind")
	if err != nil {
		return pkg, err
	}

	if !enabled {
		pkg.Config[nonlocalbindCfg] = "0"
	}

	enabled, err = s.currentSysctlConfig("net.ipv4.ip_forward")
	if err != nil {
		return pkg, err
	}

	if !enabled {
		pkg.Config[ipForwardCfg] = "0"
	}

	return pkg, nil
}

func (s *nginxDep) Ensure(remove operator.Package, ensure operator.Package) (bool, error) {

	ensureCfg, ok := ensure.Config["nginx.conf"]
	if !ok {
		s.systemd.Disable("nginx")
		os.Remove("/etc/nginx/nginx.conf")
		return false, nil
	}

	if _, ok := remove.Config["nginx.conf"]; !ok {

		if err := ioutil.WriteFile("/etc/yum.repos.d/nginx.repo", []byte(`[nginx-stable]
name=nginx stable repo
baseurl=http://nginx.org/packages/centos/$releasever/$basearch/
gpgcheck=1
enabled=1
gpgkey=https://nginx.org/keys/nginx_signing.key
module_hotfixes=true`), 0600); err != nil {
			return false, err
		}

		if err := s.manager.Install(&dep.Software{
			Package: "nginx",
			Version: ensure.Version,
		}); err != nil {
			return false, err
		}

		if err := os.MkdirAll("/etc/nginx", 0700); err != nil {
			return false, err
		}
	}

	if err := ioutil.WriteFile("/etc/nginx/nginx.conf", []byte(ensureCfg), 0600); err != nil {
		return false, err
	}

	if err := dep.ManipulateFile("/etc/sysctl.conf", []string{
		"net.ipv4.ip_forward",
		"net.ipv4.ip_nonlocal_bind",
	}, []string{
		"net.ipv4.ip_forward = 1",
		"net.ipv4.ip_nonlocal_bind = 1",
	}, nil); err != nil {
		return false, err
	}

	if err := s.systemd.Enable("nginx"); err != nil {
		return false, err
	}

	if _, ok := remove.Config[ipForwardCfg]; ok {
		return true, nil
	}

	_, ok = remove.Config[nonlocalbindCfg]
	return ok, nil
}

func (n *nginxDep) currentSysctlConfig(property string) (bool, error) {

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	cmd := exec.Command("sysctl", property)
	cmd.Stderr = &errBuf
	cmd.Stderr = &outBuf

	fullCmd := strings.Join(cmd.Args, " ")
	n.logger.WithFields(map[string]interface{}{"cmd": fullCmd}).Debug("Executing")

	if err := cmd.Run(); err != nil {
		return false, errors.Wrapf(err, "running %s failed with stderr %s", fullCmd, errBuf.String())
	}

	return outBuf.String() == fmt.Sprintf("%s = 1\n", property), nil
}
