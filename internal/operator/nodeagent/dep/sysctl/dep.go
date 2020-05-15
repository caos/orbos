package sysctl

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
	"github.com/caos/orbiter/internal/operator/nodeagent/dep/middleware"
	"github.com/caos/orbiter/mntr"
)

type Installer interface {
	isSysctl()
	nodeagent.Installer
}
type sysctlDep struct {
	monitor mntr.Monitor
}

func New(monitor mntr.Monitor) Installer {
	return &sysctlDep{monitor: monitor}
}

func (sysctlDep) isSysctl() {}

func (sysctlDep) Is(other nodeagent.Installer) bool {
	_, ok := middleware.Unwrap(other).(Installer)
	return ok
}

func (sysctlDep) String() string { return "sysctl" }

func (*sysctlDep) Equals(other nodeagent.Installer) bool {
	_, ok := other.(*sysctlDep)
	return ok
}

type SysctlPropery string

const (
	IpForward             SysctlPropery = "net.ipv4.ip_forward"
	NonLocalBind          SysctlPropery = "net.ipv4.ip_nonlocal_bind"
	BridgeNfCallIptables  SysctlPropery = "net.bridge.bridge-nf-call-iptables"
	BridgeNfCallIp6tables SysctlPropery = "net.bridge.bridge-nf-call-ip6tables"
)

func SetProperty(pkg *common.Package, propery SysctlPropery, enable bool) {

	if pkg.Config == nil {
		pkg.Config = make(map[string]string)
	}

	if _, ok := pkg.Config[string(IpForward)]; !ok {
		pkg.Config[string(IpForward)] = "0"
	}

	if _, ok := pkg.Config[string(NonLocalBind)]; !ok {
		pkg.Config[string(NonLocalBind)] = "0"
	}
	if _, ok := pkg.Config[string(BridgeNfCallIptables)]; !ok {
		pkg.Config[string(BridgeNfCallIptables)] = "0"
	}
	if _, ok := pkg.Config[string(BridgeNfCallIp6tables)]; !ok {
		pkg.Config[string(BridgeNfCallIp6tables)] = "0"
	}

	state := "0"
	if enable {
		state = "1"
	}

	pkg.Config[string(propery)] = state
}

func (s *sysctlDep) Current() (pkg common.Package, err error) {

	if err := currentSysctlConfig(s.monitor, IpForward, &pkg); err != nil {
		return pkg, err
	}

	if err := currentSysctlConfig(s.monitor, NonLocalBind, &pkg); err != nil {
		return pkg, err
	}

	if err := currentSysctlConfig(s.monitor, BridgeNfCallIptables, &pkg); err != nil {
		return pkg, err
	}

	if err := currentSysctlConfig(s.monitor, BridgeNfCallIp6tables, &pkg); err != nil {
		return pkg, err
	}

	return pkg, nil
}

func (s *sysctlDep) Ensure(_ common.Package, ensure common.Package) error {

	if err := ioutil.WriteFile("/etc/sysctl.d/30-orbiter.conf", []byte(fmt.Sprintf(
		`%s = %s
%s = %s
%s = %s
%s = %s
`,
		string(IpForward), ensure.Config[string(IpForward)],
		string(NonLocalBind), ensure.Config[string(NonLocalBind)],
		string(BridgeNfCallIptables), ensure.Config[string(BridgeNfCallIptables)],
		string(BridgeNfCallIp6tables), ensure.Config[string(BridgeNfCallIp6tables)],
	)), os.ModePerm); err != nil {
		return err
	}

	cmd := exec.Command("sysctl", "--system")
	if output, err := cmd.CombinedOutput(); err != nil {
		return errors.Wrapf(err, "running %s failed with stderr %s", strings.Join(cmd.Args, " "), string(output))
	}
	return nil
}

func currentSysctlConfig(monitor mntr.Monitor, property SysctlPropery, pkg *common.Package) error {

	propertyStr := string(property)

	var (
		outBuf bytes.Buffer
		errBuf bytes.Buffer
	)

	cmd := exec.Command("sysctl", propertyStr)
	cmd.Stderr = &errBuf
	cmd.Stdout = &outBuf

	fullCmd := strings.Join(cmd.Args, " ")
	monitor.WithFields(map[string]interface{}{"cmd": fullCmd}).Debug("Executing")

	if err := cmd.Run(); err != nil {
		errStr := errBuf.String()
		if !strings.Contains(errStr, "No such file or directory") {
			return errors.Wrapf(err, "running %s failed with stderr %s", fullCmd, errStr)
		}
	}

	if pkg.Config == nil {
		pkg.Config = make(map[string]string)
	}
	pkg.Config[propertyStr] = "0"
	enabled := outBuf.String() == fmt.Sprintf("%s = 1\n", property)
	if enabled {
		pkg.Config[propertyStr] = "1"
	}

	return nil
}
