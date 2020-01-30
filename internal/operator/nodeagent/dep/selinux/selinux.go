package selinux

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/nodeagent/dep"
	"github.com/caos/orbiter/logging"
)

func Current(os dep.OperatingSystem, pkg *common.Package) (err error) {

	if os != dep.CentOS {
		return nil
	}

	var buf bytes.Buffer
	cmd := exec.Command("sestatus")
	cmd.Stdout = &buf
	if err := cmd.Run(); err != nil {
		return err
	}

	for err == nil {
		line, err := buf.ReadString('\n')
		if err != nil {
			return err
		}
		if strings.Contains(line, "Current mode:") {
			status := strings.TrimSpace(strings.Split(line, ":")[1])
			if status != "permissive" {
				pkg.Config["selinux"] = status
				return nil
			}
		}
	}
	return err
}

func EnsurePermissive(logger logging.Logger, opsys dep.OperatingSystem, remove common.Package) error {

	if opsys != dep.CentOS || remove.Config["selinux"] == "permissive" {
		return nil
	}

	var errBuf bytes.Buffer
	cmd := exec.Command("setenforce", "0")
	cmd.Stderr = &errBuf
	if logger.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "disabling SELinux while installing kubelet so that containers can access the host filesystem failed with stderr %s", errBuf.String())
	}
	errBuf.Reset()

	cmd = exec.Command("sed", "-i", "s/^SELINUX=enforcing$/SELINUX=permissive/", "/etc/selinux/config")
	cmd.Stderr = &errBuf
	if logger.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "disabling SELinux while installing kubelet so that containers can access the host filesystem failed with stderr %s", errBuf.String())
	}
	return nil
}
