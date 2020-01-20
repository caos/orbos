package dep

import (
	"fmt"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/logging"
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
	logger    logging.Logger
	os        OperatingSystem
	installed map[string]string
}

func (p *PackageManager) RefreshInstalled() error {
	var err error
	switch p.os.Packages {
	case DebianBased:
		err = p.debbasedInstalled()
	case REMBased:
		err = p.rembasedInstalled()
	}

	p.logger.WithFields(map[string]interface{}{
		"packages": len(p.installed),
	}).Debug("Refreshed installed packages")

	return errors.Wrap(err, "refreshing installed packages failed")
}

func (p *PackageManager) Init() error {

	p.logger.Info("Updating packages")

	var err error
	switch p.os.Packages {
	case DebianBased:
		err = p.debSpecificUpdatePackages()
	case REMBased:
		err = p.remSpecificUpdatePackages()
	}

	if err != nil {
		return errors.Wrapf(err, "updating packages failed", p.os.Packages)
	}

	p.logger.Info("Packages are updated")
	return nil
}

func NewPackageManager(logger logging.Logger, os OperatingSystem) *PackageManager {
	return &PackageManager{logger, os, nil}
}

func (p *PackageManager) CurrentVersions(possiblePackages ...string) ([]*Software, error) {

	software := make([]*Software, 0)
	for _, pkg := range possiblePackages {
		if version, ok := p.installed[pkg]; ok {
			pkg := &Software{
				Package: pkg,
				Version: version,
			}
			software = append(software, pkg)
			p.logger.WithFields(map[string]interface{}{
				"package": pkg.Package,
				"version": pkg.Version,
			}).Debug("Found filtered installed package")
		}
	}

	return software, nil
}

func (p *PackageManager) Install(installVersion *Software, more ...*Software) error {
	switch p.os.Packages {
	case DebianBased:
		return p.debbasedInstall(installVersion, more...)
	case REMBased:
		return p.rembasedInstall(installVersion, more...)
	}
	return errors.Errorf("Package manager %s is not implemented", p.os.Packages)
}

func (p *PackageManager) Add(repo *Repository) error {
	switch p.os.Packages {
	case DebianBased:
		return p.debbasedAdd(repo)
	case REMBased:
		return p.rembasedAdd(repo)
	}
	return errors.Errorf("Package manager %s is not implemented", p.os.Packages)
}
