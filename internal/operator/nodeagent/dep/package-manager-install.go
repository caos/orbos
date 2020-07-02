package dep

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
)

func (p *PackageManager) rembasedInstall(installVersion *Software, more ...*Software) error {

	if err := os.MkdirAll("/etc/yum/pluginconf.d", 0700); err != nil {
		return err
	}

	if err := touch("/etc/yum/pluginconf.d/versionlock.list"); err != nil {
		return err
	}

	pkgs := make([]string, len(more)+1)
	for idx, sw := range append([]*Software{installVersion}, more...) {
		pkgs[idx] = sw.Package
		if sw.Version == "" {
			continue
		}

		installFmt := fmt.Sprintf("%s-%s", sw.Package, sw.Version)
		if err := ManipulateFile("/etc/yum/pluginconf.d/versionlock.list", []string{installFmt}, []string{fmt.Sprintf("%s.*", installFmt)}, nil); err != nil {
			return err
		}
	}

	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()

	cmd := exec.Command("yum", append([]string{"install", "-y"}, pkgs...)...)
	cmd.Stderr = errBuf
	if p.monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	if err := errors.Wrapf(cmd.Run(), "installing yum packages %s failed with stderr %s", strings.Join(pkgs, " and "), errBuf.String()); err != nil {
		return err
	}
	//	errBuf.Reset()
	//
	//	cmd := exec.Command("yum", append([]string{"downgrade", "-y"}, pkgs...)...)
	//	cmd.Stderr = errBuf
	//	if p.monitor.IsVerbose() {
	//		fmt.Println(strings.Join(cmd.Args, " "))
	//		cmd.Stdout = os.Stdout
	//	}
	//	err := cmd.Run()
	//	stdErr := errBuf.String()
	// TODO: Already installed is no error
	//	if err != nil {
	//		return errors.Wrapf(err, "installing yum packages %s failed with stderr %s", strings.Join(pkgs, " and "), stdErr)
	//	}
	return nil
}

// TODO: Use lower level apt instead of apt-get?
func (p *PackageManager) debbasedInstall(installVersion *Software, more ...*Software) error {

	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()

	pkgs := make([]string, len(more)+1)
	hold := make([]string, 0)
	for idx, sw := range append([]*Software{installVersion}, more...) {
		pkgs[idx] = sw.Package
		if sw.Version == "" {
			continue
		}
		pkgs[idx] = fmt.Sprintf("%s=%s", sw.Package, sw.Version)
		hold = append(hold, sw.Package)

		cmd := exec.Command("apt-mark", "unhold", sw.Package)
		cmd.Stderr = errBuf
		if p.monitor.IsVerbose() {
			fmt.Println(strings.Join(cmd.Args, " "))
			cmd.Stdout = os.Stdout
		}
		if err := cmd.Run(); err != nil {
			return errors.Wrapf(err, "unholding installed package failed with stderr %s", errBuf.String())
		}
		errBuf.Reset()
	}

	cmd := exec.Command("dpkg", "--configure", "-a")
	cmd.Stderr = errBuf
	if p.monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "cleaning up dpkg failed with stderr %s", errBuf.String())
	}
	errBuf.Reset()

	cmd = exec.Command("apt-get", append(strings.Fields(
		"--assume-yes --allow-downgrades install -y"), pkgs...)...)
	cmd.Stderr = errBuf
	if p.monitor.IsVerbose() {
		fmt.Println(strings.Join(cmd.Args, " "))
		cmd.Stdout = os.Stdout
	}
	if err := cmd.Run(); err != nil {
		return errors.Wrapf(err, "installing package failed with stderr %s", errBuf.String())
	}
	errBuf.Reset()

	for _, pkg := range hold {
		cmd = exec.Command("apt-mark", "hold", pkg)
		cmd.Stderr = errBuf
		if p.monitor.IsVerbose() {
			fmt.Println(strings.Join(cmd.Args, " "))
			cmd.Stdout = os.Stdout
		}
		if err := cmd.Run(); err != nil {
			return errors.Wrapf(err, "holding package failed with stderr %s", errBuf.String())
		}
		errBuf.Reset()

		p.monitor.WithFields(map[string]interface{}{
			"package": installVersion.Package,
			"version": installVersion.Version,
		}).Debug("Installed package")
	}
	return nil
}

func touch(fileName string) error {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		file, err := os.Create("temp.txt")
		if err != nil {
			return err
		}
		defer file.Close()
	} else {
		currentTime := time.Now().Local()
		err = os.Chtimes(fileName, currentTime, currentTime)
		if err != nil {
			return err
		}
	}
	return nil
}
