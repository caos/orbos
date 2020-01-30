package keepalived

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/nodeagent"
	"github.com/caos/orbiter/internal/operator/nodeagent/dep"
	"github.com/caos/orbiter/internal/operator/nodeagent/dep/middleware"
	"github.com/caos/orbiter/internal/operator/nodeagent/dep/selinux"
	"github.com/caos/orbiter/logging"
)

type Installer interface {
	isKeepalived()
	nodeagent.Installer
}
type keepaliveDDep struct {
	logger   logging.Logger
	manager  *dep.PackageManager
	systemd  *dep.SystemD
	peerAuth string
	os       dep.OperatingSystem
}

func New(logger logging.Logger, manager *dep.PackageManager, systemd *dep.SystemD, os dep.OperatingSystem, cipher string) Installer {
	return &keepaliveDDep{logger, manager, systemd, cipher[:8], os}
}

func (keepaliveDDep) isKeepalived() {}

func (keepaliveDDep) Is(other nodeagent.Installer) bool {
	_, ok := middleware.Unwrap(other).(Installer)
	return ok
}

func (keepaliveDDep) String() string { return "Keepalived" }

func (*keepaliveDDep) Equals(other nodeagent.Installer) bool {
	_, ok := other.(*keepaliveDDep)
	return ok
}

const (
	ipForwardCfg    = "/proc/sys/net/ipv4/ip_forward"
	nonlocalbindCfg = "/proc/sys/net/ipv4/ip_nonlocal_bind"
)

func (s *keepaliveDDep) Current() (pkg common.Package, err error) {
	defer func() {
		if err != nil {
			err = selinux.Current(s.os, &pkg)
		}
	}()
	config, err := ioutil.ReadFile("/etc/keepalived/keepalived.conf")
	if os.IsNotExist(err) {
		return pkg, nil
	}

	var redacted bytes.Buffer
	dep.Manipulate(bytes.NewReader(config), &redacted, nil, nil, func(line string) *string {
		searchString := "auth_pass "
		if strings.Contains(line, searchString) {
			line = line[0:strings.Index(line, searchString)+len(searchString)] + "[ REDACTED ]"
		}
		return &line
	})
	pkg.Config = map[string]string{
		"keepalived.conf": redacted.String(),
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

	notifymaster, err := ioutil.ReadFile("/etc/keepalived/notifymaster.sh")
	if os.IsNotExist(err) {
		return pkg, nil
	}
	pkg.Config["notifymaster.sh"] = string(notifymaster)
	return pkg, err
}

func (s *keepaliveDDep) Ensure(remove common.Package, ensure common.Package) (bool, error) {

	if err := selinux.EnsurePermissive(s.logger, s.os, remove); err != nil {
		return false, err
	}

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

	if notifyMaster, ok := ensure.Config["notifymaster.sh"]; ok {
		if err := ioutil.WriteFile("/etc/keepalived/notifymaster.sh", []byte(notifyMaster), 0777); err != nil {
			return false, err
		}
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

	if err := s.systemd.Enable("keepalived"); err != nil {
		return false, err
	}
	if _, ok := remove.Config[ipForwardCfg]; ok {
		return true, nil
	}

	if _, ok = remove.Config[nonlocalbindCfg]; ok {
		return true, nil
	}

	return false, s.systemd.Start("keepalived")
}

func (k *keepaliveDDep) currentSysctlConfig(property string) (bool, error) {

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	cmd := exec.Command("sysctl", property)
	cmd.Stderr = &errBuf
	cmd.Stdout = &outBuf

	fullCmd := strings.Join(cmd.Args, " ")
	k.logger.WithFields(map[string]interface{}{"cmd": fullCmd}).Debug("Executing")

	if err := cmd.Run(); err != nil {
		return false, errors.Wrapf(err, "running %s failed with stderr %s", fullCmd, errBuf.String())
	}

	return outBuf.String() == fmt.Sprintf("%s = 1\n", property), nil
}
