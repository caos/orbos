package keepalived

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/nodeagent"
	"github.com/caos/orbos/internal/operator/nodeagent/dep"
	"github.com/caos/orbos/internal/operator/nodeagent/dep/middleware"
	"github.com/caos/orbos/internal/operator/nodeagent/dep/selinux"
	"github.com/caos/orbos/mntr"
)

type Installer interface {
	isKeepalived()
	nodeagent.Installer
}
type keepaliveDDep struct {
	monitor  mntr.Monitor
	manager  *dep.PackageManager
	systemd  *dep.SystemD
	peerAuth string
	os       dep.OperatingSystem
}

func New(monitor mntr.Monitor, manager *dep.PackageManager, systemd *dep.SystemD, os dep.OperatingSystem, cipher string) Installer {
	return &keepaliveDDep{monitor, manager, systemd, cipher[:8], os}
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

func (s *keepaliveDDep) Current() (pkg common.Package, err error) {
	if !s.systemd.Active("keepalived") {
		return pkg, err
	}

	defer func() {
		if err == nil {
			err = selinux.Current(s.os, &pkg)
		}
	}()
	config, err := ioutil.ReadFile("/etc/keepalived/keepalived.conf")
	if os.IsNotExist(err) {
		return pkg, nil
	}

	redacted := new(bytes.Buffer)
	defer redacted.Reset()

	dep.Manipulate(bytes.NewReader(config), redacted, nil, nil, func(line string) *string {
		searchString := "auth_pass "
		if strings.Contains(line, searchString) {
			line = line[0:strings.Index(line, searchString)+len(searchString)] + "[ REDACTED ]"
		}
		return &line
	})
	pkg.Config = map[string]string{
		"keepalived.conf": redacted.String(),
	}

	notifymaster, err := ioutil.ReadFile("/etc/keepalived/notifymaster.sh")
	if os.IsNotExist(err) {
		err = nil
	}
	if err != nil {
		return pkg, err
	}

	if string(notifymaster) != "" {
		pkg.Config["notifymaster.sh"] = string(notifymaster)
	}

	authCheck, err := ioutil.ReadFile("/etc/keepalived/authcheck.sh")
	if os.IsNotExist(err) {
		err = nil
	}
	if err != nil {
		return pkg, err
	}
	if string(authCheck) != "" {
		pkg.Config["authcheck.sh"] = string(authCheck)
		pkg.Config["authcheckexitcode"] = strconv.Itoa(exec.Command("/etc/keepalived/authcheck.sh").Run().(*exec.ExitError).ExitCode())
	}

	return pkg, err
}

func (s *keepaliveDDep) Ensure(remove common.Package, ensure common.Package) error {

	if err := selinux.EnsurePermissive(s.monitor, s.os, remove); err != nil {
		return err
	}

	ensureCfg, ok := ensure.Config["keepalived.conf"]
	if !ok {
		if err := s.systemd.Disable("keepalived"); err != nil {
			return err
		}
		return os.Remove("/etc/keepalived/keepalived.conf")
	}

	if err := s.manager.Install(&dep.Software{
		Package: "keepalived",
		Version: ensure.Version,
	}); err != nil {
		return err
	}

	if err := os.MkdirAll("/etc/keepalived", 0700); err != nil {
		return err
	}

	if err := ioutil.WriteFile("/etc/keepalived/keepalived.conf", []byte(strings.ReplaceAll(ensureCfg, "[ REDACTED ]", s.peerAuth)), 0600); err != nil {
		return err
	}

	if notifyMaster, ok := ensure.Config["notifymaster.sh"]; ok {
		if err := ioutil.WriteFile("/etc/keepalived/notifymaster.sh", []byte(notifyMaster), 0777); err != nil {
			return err
		}
	}

	if err := s.systemd.Enable("keepalived"); err != nil {
		return err
	}

	return s.systemd.Start("keepalived")
}
