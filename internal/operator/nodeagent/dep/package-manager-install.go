package dep

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/caos/orbos/mntr"
)

func (p *PackageManager) rembasedInstall(install ...*Software) error {

	if len(install) == 0 {
		return nil
	}

	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()

	installPkgs := make([]string, 0)
	for _, sw := range install {

		_, ok := p.installed[sw.Package]
		if ok && sw.Version == "" {
			continue
		}

		if sw.Version == "" {
			installPkgs = append(installPkgs, sw.Package)
			continue
		}

		installPkg := fmt.Sprintf("%s-%s", sw.Package, sw.Version)
		installPkgs = append(installPkgs, installPkg)
		cmd := exec.Command("yum", "versionlock", "delete", sw.Package)
		cmd.Stderr = errBuf
		if p.monitor.IsVerbose() {
			fmt.Println(strings.Join(cmd.Args, " "))
			cmd.Stdout = os.Stdout
		}
		err := cmd.Run()
		stderr := errBuf.String()
		if err != nil && !strings.Contains(stderr, "versionlock delete: no matches") {
			return fmt.Errorf("unlocking package %s failed with stderr %s: %w", sw.Package, stderr, err)
		}
		errBuf.Reset()

		cmd = exec.Command("yum", "versionlock", "add", "-y", installPkg)
		cmd.Stderr = errBuf
		if p.monitor.IsVerbose() {
			fmt.Println(strings.Join(cmd.Args, " "))
			cmd.Stdout = os.Stdout
		}
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("locking package %s at version %s failed with stderr %s: %w", sw.Package, sw.Version, errBuf.String(), err)
		}
		errBuf.Reset()
	}

	for _, pkg := range installPkgs {
		if err := rembasedInstallPkg(p.monitor, pkg); err != nil {
			return err
		}
	}
	return nil
}

func rembasedInstallPkg(monitor mntr.Monitor, pkg string) error {
	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()
	outBuf := new(bytes.Buffer)
	defer outBuf.Reset()
	cmd := exec.Command("yum", "install", "-y", pkg)
	cmd.Stderr = errBuf
	cmd.Stdout = outBuf
	err := cmd.Run()
	errStr := errBuf.String()
	outStr := outBuf.String()
	monitor.WithFields(map[string]interface{}{
		"command": fmt.Sprintf("'%s'", strings.Join(cmd.Args, "' '")),
		"stdout":  outStr,
		"stderr":  errStr,
	}).Debug("Executed yum install")
	if err != nil && !strings.Contains(errStr+outStr, "is already installed") {
		return fmt.Errorf("installing yum package %s failed with stderr %s: %w", pkg, errStr, err)
	}
	return nil
}

// TODO: Use lower level apt instead of apt-get?
func (p *PackageManager) debbasedInstall(install ...*Software) error {

	errBuf := new(bytes.Buffer)
	defer errBuf.Reset()

	pkgs := make([]string, len(install))
	hold := make([]string, 0)
	for idx, sw := range install {
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
			return fmt.Errorf("unholding installed package failed with stderr %s: %w", errBuf.String(), err)
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
		return fmt.Errorf("cleaning up dpkg failed with stderr %s: %w", errBuf.String(), err)
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
		return fmt.Errorf("installing package failed with stderr %s: %w", errBuf.String(), err)
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
			return fmt.Errorf("holding package failed with stderr %s: %w", errBuf.String(), err)
		}
		errBuf.Reset()

		p.monitor.WithFields(map[string]interface{}{
			"software": pkg,
		}).Debug("Holded package")
	}
	return nil
}
