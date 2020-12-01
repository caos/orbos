package dep

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"

	"github.com/pkg/errors"
)

func (p *PackageManager) debSpecificUpdatePackages() error {
	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()
	cmd := exec.Command("apt-get", "--assume-yes", "update")
	cmd.Stderr = errBuf
	if p.monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "updating deb packages failed with stderr %s", errBuf.String())
	}

	return p.debbasedInstall(
		&Software{Package: "apt-transport-https"},
		&Software{Package: "gnupg2"},
		&Software{Package: "software-properties-common"})
}

func (p *PackageManager) remSpecificUpdatePackages() error {

	if err := ioutil.WriteFile("/etc/cron.daily/yumupdate.sh", []byte(`#!/bin/bash
YUM=/usr/bin/yum
$YUM -y -R 10 -e 3 -d 3 update
`), 0777); err != nil {
		return err
	}

	if err := p.rembasedInstall(
		&Software{Package: "yum-utils"},
		&Software{Package: "yum-plugin-versionlock"},
		&Software{Package: "firewalld"},
	); err != nil {
		return err
	}

	return p.systemd.Enable("firewalld")
}
