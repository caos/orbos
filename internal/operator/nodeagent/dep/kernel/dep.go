package kernel

import (
	"context"
	"fmt"
	"io/fs"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/caos/orbos/mntr"

	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/nodeagent"
	"github.com/caos/orbos/internal/operator/nodeagent/dep"
	"github.com/caos/orbos/internal/operator/nodeagent/dep/middleware"
)

var _ Installer = (*kernelDep)(nil)

type Installer interface {
	isKernel()
	nodeagent.Installer
}

type kernelDep struct {
	ctx     context.Context
	monitor mntr.Monitor
	manager *dep.PackageManager
}

/*
New returns the kernel dependency

Node Agent uninstalls all kernels that don't have a corresponding initramfs file except for the currently
loaded kernel. If the currently loaded kernel doesn't have a corresponding initramfs file, Node Agent panics.

If ORBITER desires a specific kernel version, Node Agent installs and locks it, checks the initramfs file and reboots.
It is in the ORBITERS responsibility to ensure not all nodes are updated and rebooted simultaneously.
*/
func New(ctx context.Context, monitor mntr.Monitor, manager *dep.PackageManager) *kernelDep {
	return &kernelDep{
		ctx:     ctx,
		monitor: monitor,
		manager: manager,
	}
}

func (kernelDep) isKernel() {}

func (kernelDep) Is(other nodeagent.Installer) bool {
	_, ok := middleware.Unwrap(other).(Installer)
	return ok
}

func (kernelDep) String() string { return "Kernel" }

func (*kernelDep) Equals(other nodeagent.Installer) bool {
	_, ok := other.(*kernelDep)
	return ok
}

func (k *kernelDep) Current() (pkg common.Package, err error) {
	loaded, corrupted, err := k.kernelVersions()
	if err != nil {
		return pkg, err
	}

	dashIdx := strings.Index(loaded, "-")
	pkg.Version = loaded[0:dashIdx]
	pkg.Config = map[string]string{
		dep.CentOS7.String(): loaded[dashIdx+1:],
	}

	if len(corrupted) > 0 {
		pkg.Config = map[string]string{"corrupted": strings.Join(corrupted, ",")}
	}

	return pkg, nil
}

func (k *kernelDep) Ensure(remove common.Package, ensure common.Package) error {

	corruptedKernels := make([]*dep.Software, 0)
	corruptedKernelsStr, ok := remove.Config["corrupted"]
	if ok && corruptedKernelsStr != "" {
		corruptedKernelsStrs := strings.Split(corruptedKernelsStr, ",")
		for i := range corruptedKernelsStrs {
			corruptedKernels = append(corruptedKernels, &dep.Software{
				Package: "kernel",
				Version: corruptedKernelsStrs[i],
			})
		}
	}

	if err := k.manager.Remove(corruptedKernels...); err != nil {
		return err
	}

	removeVersion := fmt.Sprintf("%s-%s", ensure.Version, ensure.Config[dep.CentOS7.String()])
	ensureVersion := fmt.Sprintf("%s-%s", ensure.Version, ensure.Config[dep.CentOS7.String()])

	if removeVersion == ensureVersion || ensure.Version == "" {
		return nil
	}

	if err := k.manager.Install(&dep.Software{
		Package: "kernel",
		Version: ensureVersion,
	}); err != nil {
		return err
	}

	initramfsVersions, err := listInitramfsVersions()
	if err != nil {
		return err
	}

	var found bool
	for i := range initramfsVersions {
		if initramfsVersions[i] == ensure.Version {
			found = true
			break
		}
	}
	if !found {
		return fmt.Errorf("couldn't find a corresponding initramfs file corresponding kernel version %s. Not rebooting", ensure.Version)
	}

	k.monitor.Info("Stopping node-agentd in the background")
	if err := exec.Command("bash", "-c", "systemctl stop node-agentd &").Run(); err != nil {
		return fmt.Errorf("stopping node-agentd failed: %w", err)
	}

	k.monitor.Info("Rebooting system in the background")
	if err := exec.Command("bash", "-c", "reboot &").Run(); err != nil {
		return fmt.Errorf("rebooting the system failed: %w", err)
	}

	k.monitor.Info("Awaiting SIGTERM")
	<-k.ctx.Done()
	return nil
}

func (k *kernelDep) kernelVersions() (loadedKernel string, corruptedKernels []string, err error) {

	loadedKernelBytes, err := exec.Command("uname", "-r").CombinedOutput()
	if err != nil {
		return loadedKernel, corruptedKernels, fmt.Errorf("getting loaded kernel failed: %s: %w", loadedKernel, err)
	}

	loadedKernel = trimArchitecture(string(loadedKernelBytes))

	initramfsVersions, err := listInitramfsVersions()
	if err != nil {
		return loadedKernel, corruptedKernels, err
	}

	corruptedKernels = make([]string, 0)
kernels:
	for _, installedKernel := range k.manager.CurrentVersions("kernel") {
		for i := range initramfsVersions {
			if initramfsVersions[i] == installedKernel.Version {
				continue kernels
			}
		}
		if installedKernel.Version == loadedKernel {
			panic("The actively loaded kernel has no corresponding initramfs file. Pleases fix it manually so the machine survives the next reboot")
		}
		corruptedKernels = append(corruptedKernels, installedKernel.Version)
	}

	return loadedKernel, corruptedKernels, nil
}

func listInitramfsVersions() ([]string, error) {
	initramfsdir := "/boot/"

	var initramfsKernels []string
	if err := filepath.WalkDir(initramfsdir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path != initramfsdir && d.IsDir() {
			return filepath.SkipDir
		}
		if strings.HasPrefix(path, initramfsdir+"initramfs-") && strings.HasSuffix(path, ".img") {
			version := trimArchitecture(path[len(initramfsdir+"initramfs-"):strings.LastIndex(path, ".img")])
			initramfsKernels = append(initramfsKernels, version)
		}
		return nil
	}); err != nil {
		return nil, err
	}
	return initramfsKernels, nil
}

func trimArchitecture(kernel string) string {
	return strings.TrimSuffix(strings.TrimSuffix(kernel, "\n"), ".x86_64")
}
