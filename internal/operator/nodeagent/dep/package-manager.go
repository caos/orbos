package dep

import (
	"fmt"

	"github.com/caos/orbos/v5/mntr"
)

type Software struct {
	Package string
	Version string
}

func (s *Software) String() string {
	return fmt.Sprintf("%s=%s", s.Package, s.Version)
}

type Repository struct {
	Repository     string
	KeyURL         string
	KeyFingerprint string
}

type PackageManager struct {
	monitor   mntr.Monitor
	os        OperatingSystem
	installed map[string]string
	systemd   *SystemD
}

func (p *PackageManager) RefreshInstalled() error {
	var err error
	switch p.os.Packages {
	case DebianBased:
		err = p.debbasedInstalled()
	case REMBased:
		err = p.rembasedInstalled()
	}

	if err != nil {
		return fmt.Errorf("refreshing installed packages failed: %w", err)
	}
	p.monitor.WithFields(map[string]interface{}{
		"packages": len(p.installed),
	}).Debug("Refreshed installed packages")
	return nil
}

func (p *PackageManager) Init() error {

	p.monitor.Debug("Initializing package manager")
	var err error
	switch p.os.Packages {
	case DebianBased:
		err = p.debSpecificInit()
	case REMBased:
		err = p.remSpecificInit()
	}

	if err != nil {
		return fmt.Errorf("initializing packages %s failed: %w", p.os.Packages, err)
	}

	p.monitor.Debug("Package manager initialized")
	return nil
}

func (p *PackageManager) Update() error {
	p.monitor.Debug("Updating packages")
	var err error
	switch p.os.Packages {
	case DebianBased:
		err = p.debSpecificUpdatePackages()
	case REMBased:
		err = p.remSpecificUpdatePackages()
	}

	if err != nil {
		return fmt.Errorf("updating packages %s failed: %w", p.os.Packages, err)
	}

	p.monitor.Info("Packages updated")
	return nil
}

func NewPackageManager(monitor mntr.Monitor, os OperatingSystem, systemd *SystemD) *PackageManager {
	return &PackageManager{monitor, os, nil, systemd}
}

func (p *PackageManager) CurrentVersions(possiblePackages ...string) []*Software {

	software := make([]*Software, 0)
	for _, pkg := range possiblePackages {
		if version, ok := p.installed[pkg]; ok {
			pkg := &Software{
				Package: pkg,
				Version: version,
			}
			software = append(software, pkg)
			p.monitor.WithFields(map[string]interface{}{
				"package": pkg.Package,
				"version": pkg.Version,
			}).Debug("Found filtered installed package")
		}
	}

	return software
}

func (p *PackageManager) Install(installVersion *Software, more ...*Software) error {
	switch p.os.Packages {
	case DebianBased:
		return p.debbasedInstall(installVersion, more...)
	case REMBased:
		return p.rembasedInstall(installVersion, more...)
	}
	return fmt.Errorf("package manager %s is not implemented", p.os.Packages)
}

func (p *PackageManager) Add(repo *Repository) error {
	switch p.os.Packages {
	case DebianBased:
		return p.debbasedAdd(repo)
	case REMBased:
		return p.rembasedAdd(repo)
	default:
		return fmt.Errorf("package manager %s is not implemented", p.os.Packages)
	}
}
