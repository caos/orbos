package kubelet

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/nodeagent"
	"github.com/caos/orbiter/internal/operator/nodeagent/dep"
	"github.com/caos/orbiter/internal/operator/nodeagent/dep/k8s"
	"github.com/caos/orbiter/internal/operator/nodeagent/dep/middleware"
	"github.com/caos/orbiter/internal/operator/nodeagent/dep/selinux"
	"github.com/caos/orbiter/mntr"
)

type Installer interface {
	isKubelet()
	nodeagent.Installer
}

type kubeletDep struct {
	os      dep.OperatingSystem
	monitor mntr.Monitor
	common  *k8s.Common
	systemd *dep.SystemD
}

func New(monitor mntr.Monitor, os dep.OperatingSystem, manager *dep.PackageManager, systemd *dep.SystemD) Installer {
	return &kubeletDep{os, monitor, k8s.New(os, manager, "kubelet"), systemd}
}

func (kubeletDep) isKubelet() {}

func (kubeletDep) Is(other nodeagent.Installer) bool {
	_, ok := middleware.Unwrap(other).(Installer)
	return ok
}

func (k kubeletDep) String() string { return "Kubelet" }

func (*kubeletDep) Equals(other nodeagent.Installer) bool {
	_, ok := other.(*kubeletDep)
	return ok
}

const (
	ipForwardCfg = "net.ipv4.ip_forward"
	iptables     = "net.bridge.bridge-nf-call-iptables"
	ip6tables    = "net.bridge.bridge-nf-call-ip6tables"
)

func (k *kubeletDep) Current() (common.Package, error) {
	pkg, err := k.common.Current()
	if err != nil {
		return pkg, err
	}

	enabled, err := dep.CurrentSysctlConfig(k.monitor, ipForwardCfg)
	if err != nil {
		return pkg, err
	}

	if !enabled {
		pkg.Config[ipForwardCfg] = "0"
	}

	enabled, err = dep.CurrentSysctlConfig(k.monitor, iptables)
	if err != nil {
		return pkg, err
	}

	if !enabled {
		pkg.Config[iptables] = "0"
	}

	enabled, err = dep.CurrentSysctlConfig(k.monitor, ip6tables)
	if err != nil {
		return pkg, err
	}

	if !enabled {
		pkg.Config[ip6tables] = "0"
	}

	return pkg, selinux.Current(k.os, &pkg)
}

func (k *kubeletDep) Ensure(remove common.Package, install common.Package) error {

	if err := selinux.EnsurePermissive(k.monitor, k.os, remove); err != nil {
		return err
	}

	if k.os != dep.CentOS {
		return k.ensurePackage(remove, install)
	}

	var errBuf bytes.Buffer
	cmd := exec.Command("modprobe", "br_netfilter")
	cmd.Stderr = &errBuf
	if k.monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "loading module br_netfilter while installing kubelet failed with stderr %s", errBuf.String())
	}
	errBuf.Reset()

	file, err := os.Create("/etc/sysctl.d/22-k8s.conf")
	if err != nil {
		return errors.Wrap(err, "opening /etc/sysctl.d/22-k8s.conf failed")
	}
	defer file.Close()

	file.Write([]byte(fmt.Sprintf(`%s = 1
%s = 1
%s = 1
		`, ipForwardCfg, ip6tables, iptables)))
	if err != nil {
		return errors.Wrap(err, "writing to /etc/sysctl.d/22-k8s.conf failed")
	}
	file.Close()

	cmd = exec.Command("sysctl", "--system")
	cmd.Stderr = &errBuf
	if k.monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}

	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "running sysctl --system failed with stderr %s", errBuf.String())
	}

	return k.ensurePackage(remove, install)
}

func (k *kubeletDep) ensurePackage(remove common.Package, install common.Package) error {
	if err := k.common.Ensure(remove, install); err != nil {
		return err
	}

	if err := k.systemd.Enable("kubelet"); err != nil {
		return err
	}

	return k.systemd.Start("kubelet")
}
