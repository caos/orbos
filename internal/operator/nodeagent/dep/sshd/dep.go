package sshd

import (
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/nodeagent"
	"github.com/caos/orbiter/internal/operator/nodeagent/dep"
	"github.com/caos/orbiter/internal/operator/nodeagent/dep/middleware"
)

type Installer interface {
	isSSHD()
	nodeagent.Installer
}

type sshdDep struct {
	systemd *dep.SystemD
}

func New(systemd *dep.SystemD) Installer {
	return &sshdDep{systemd}
}

func (sshdDep) Is(other nodeagent.Installer) bool {
	_, ok := middleware.Unwrap(other).(Installer)
	return ok
}

func (sshdDep) isSSHD() {}

func (sshdDep) String() string { return "SSHD" }

func (*sshdDep) Equals(other nodeagent.Installer) bool {
	_, ok := other.(*sshdDep)
	return ok
}

func (s *sshdDep) Current() (pkg common.Package, err error) {

	var buf bytes.Buffer
	swapon := exec.Command("sshd", "-T")
	swapon.Stdout = &buf
	if err := swapon.Run(); err != nil {
		return pkg, err
	}

	for {
		if err != nil && err != io.EOF {
			return pkg, err
		}
		line, err := buf.ReadString('\n')
		if strings.Contains(line, "listenaddress") {
			fields := strings.Fields(line)
			value := ""
			if len(fields) > 1 {
				value = fields[1]
			}
			if value != "[::]:22" && value != "0.0.0.0:22" {
				if pkg.Config == nil {
					pkg.Config = make(map[string]string)
				}
				pkg.Config["listenaddress"] = value
			}
			return pkg, nil
		}
		if err == io.EOF {
			return pkg, nil
		}
	}
}

func (s *sshdDep) Ensure(remove common.Package, ensure common.Package) (bool, error) {

	var appendLines []string
	listenAddress := ensure.Config["listenaddress"]
	if listenAddress != "" {
		appendLines = append(appendLines, fmt.Sprintf("ListenAddress %s", listenAddress))
	}

	if err := dep.ManipulateFile("/etc/ssh/sshd_config", nil, appendLines, func(line string) string {
		if strings.HasPrefix(line, "ListenAddress") {
			return ""
		}
		return line
	}); err != nil {
		return false, err
	}

	return false, s.systemd.Start("sshd")
}
