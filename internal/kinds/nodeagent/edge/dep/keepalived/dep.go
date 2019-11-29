package keepalived

import (
	"bytes"
	"io/ioutil"
	"os"
	"strings"

	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/kinds/nodeagent/adapter"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/dep"
	"github.com/caos/orbiter/internal/kinds/nodeagent/edge/dep/middleware"
)

type Installer interface {
	isKeepalived()
	adapter.Installer
}
type keepaliveDDep struct {
	manager  *dep.PackageManager
	systemd  *dep.SystemD
	peerAuth string
}

func New(manager *dep.PackageManager, systemd *dep.SystemD, cipher string) Installer {
	return &keepaliveDDep{manager, systemd, cipher[:8]}
}

func (keepaliveDDep) isKeepalived() {}

func (keepaliveDDep) Is(other adapter.Installer) bool {
	_, ok := middleware.Unwrap(other).(Installer)
	return ok
}

func (keepaliveDDep) String() string { return "Keepalived" }

func (*keepaliveDDep) Equals(other adapter.Installer) bool {
	_, ok := other.(*keepaliveDDep)
	return ok
}

const (
	ipForwardCfg    = "/proc/sys/net/ipv4/ip_forward"
	nonlocalbindCfg = "/proc/sys/net/ipv4/ip_nonlocal_bind"
)

func (s *keepaliveDDep) Current() (pkg operator.Package, err error) {
	config, err := ioutil.ReadFile("/etc/keepalived/keepalived.conf")
	if os.IsNotExist(err) {
		return pkg, nil
	}

	var redacted bytes.Buffer
	dep.Manipulate(bytes.NewReader(config), &redacted, nil, nil, func(line string) string {
		searchString := "auth_pass "
		if !strings.Contains(line, searchString) {
			return line
		}
		return line[0:strings.Index(line, searchString)+len(searchString)] + "[ REDACTED ]"
	})
	pkg.Config = map[string]string{
		"keepalived.conf": redacted.String(),
	}

	enabled, err := isEnabled(ipForwardCfg)
	if err != nil {
		return pkg, err
	}

	if !enabled {
		pkg.Config[ipForwardCfg] = "1"
	}

	enabled, err = isEnabled(nonlocalbindCfg)
	if err != nil {
		return pkg, err
	}

	if !enabled {
		pkg.Config[nonlocalbindCfg] = "1"
	}

	return pkg, err
}

func (s *keepaliveDDep) Ensure(remove operator.Package, ensure operator.Package) (bool, error) {

	ensureCfg, ok := ensure.Config["keepalived.conf"]
	if !ok {
		if err := s.systemd.Disable("keepalived"); err != nil {
			return false, err
		}
		return false, os.Remove("/etc/keepalived/keepalived.conf")
	}

	if err := s.manager.Install(&dep.Software{
		Package: "keepalived",
		Version: ensure.Version,
	}); err != nil {
		return false, err
	}

	if err := os.MkdirAll("/etc/keepalived", 0700); err != nil {
		return false, err
	}

	if err := ioutil.WriteFile("/etc/keepalived/keepalived.conf", []byte(strings.ReplaceAll(ensureCfg, "[ REDACTED ]", s.peerAuth)), 0600); err != nil {
		return false, err
	}

	if err := s.systemd.Enable("keepalived"); err != nil {
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

	if _, ok := remove.Config[ipForwardCfg]; ok {
		return true, nil
	}

	_, ok = remove.Config[nonlocalbindCfg]
	return ok, nil
}

func isEnabled(cfg string) (bool, error) {
	enabled, err := ioutil.ReadFile(cfg)
	if err != nil {
		return false, err
	}

	return string(enabled) == "1\n", nil
}
