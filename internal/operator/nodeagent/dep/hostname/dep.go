//go:generate goderive .

package hostname

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/caos/orbos/internal/operator/nodeagent/dep"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/nodeagent"
	"github.com/caos/orbos/internal/operator/nodeagent/dep/middleware"
)

type Installer interface {
	isHostname()
	nodeagent.Installer
}

type hostnameDep struct{}

func New() Installer {
	return &hostnameDep{}
}

func (hostnameDep) isHostname() {}

func (hostnameDep) Is(other nodeagent.Installer) bool {
	_, ok := middleware.Unwrap(other).(Installer)
	return ok
}

func (hostnameDep) String() string { return "Hostname" }

func (*hostnameDep) Equals(other nodeagent.Installer) bool {
	_, ok := other.(*hostnameDep)
	return ok
}

func (s *hostnameDep) Current() (pkg common.Package, err error) {

	buf := new(bytes.Buffer)
	defer buf.Reset()

	cmd := exec.Command("hostname")
	cmd.Stdout = buf
	if err := cmd.Run(); err != nil {
		return pkg, err
	}

	pkg.Config = map[string]string{"hostname": strings.TrimSuffix(buf.String(), "\n")}
	return pkg, nil
}

func (s *hostnameDep) Ensure(remove common.Package, ensure common.Package) error {

	oldHostname := remove.Config["hostname"]
	newHostname := ensure.Config["hostname"]
	if oldHostname == newHostname {
		return nil
	}

	buf := new(bytes.Buffer)
	defer buf.Reset()

	cmd := exec.Command("hostnamectl", "set-hostname", newHostname)
	cmd.Stdout = buf
	if err := cmd.Run(); err != nil {
		return err
	}

	comment := "# Added by node agent developed by CAOS AG"
	return dep.ManipulateFile("/etc/hosts", []string{comment}, []string{fmt.Sprintf("127.0.0.1\t%s %s", newHostname, comment)}, nil)
}
