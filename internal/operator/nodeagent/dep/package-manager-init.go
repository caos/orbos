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

	if err := ioutil.WriteFile("/etc/yum/yum-cron-hourly.conf", []byte(`
[commands]
update_cmd = default
update_messages = yes
download_updates = yes
apply_updates = yes
random_sleep = 0
[emitters]
system_name = None
emit_via = stdio
output_width = 80
[email]
email_from = root
email_to = root
email_host = localhost
[groups]
group_list = None
group_package_types = mandatory, default
[base]
debuglevel = -2
# skip_broken = True
mdpolicy = group:main
# assumeyes = True
`), 0600); err != nil {
		return err
	}

	if err := p.rembasedInstall(
		&Software{Package: "yum-utils"},
		&Software{Package: "yum-versionlock"},
		&Software{Package: "yum-cron"},
		&Software{Package: "firewalld"},
	); err != nil {
		return err
	}

	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()

	cmd := exec.Command("package-cleanup", "--cleandupes", "-y")
	cmd.Stderr = errBuf
	if p.monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "cleaning up duplicates failed with stderr %s", errBuf.String())
	}

	for _, unit := range []string{"yum-cron", "firewalld"} {
		if err := p.systemd.Enable(unit); err != nil {
			return err
		}
	}
	return nil
}
