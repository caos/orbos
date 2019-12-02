package middleware

import (
	"github.com/pkg/errors"

	"github.com/caos/orbiter/logging"
	"github.com/caos/orbiter/internal/core/operator"
	"github.com/caos/orbiter/internal/kinds/nodeagent/adapter"
)

type loggedDep struct {
	logger logging.Logger
	*wrapped
	unwrapped adapter.Installer
}

func AddLogging(logger logging.Logger, original adapter.Installer) Installer {
	return &loggedDep{
		logger.WithFields(map[string]interface{}{
			"dependency": original,
		}),
		&wrapped{original},
		Unwrap(original),
	}
}

func (l *loggedDep) Current() (operator.Package, error) {
	current, err := l.unwrapped.Current()
	if err == nil {
		l.logger.WithFields(map[string]interface{}{
			"version": current,
		}).Debug("Queried current dependency version")
	}
	return current, errors.Wrapf(err, "querying installed package for dependency %s failed", l.String())
}

func (l *loggedDep) Ensure(remove operator.Package, install operator.Package) (bool, error) {
	reboot, err := l.unwrapped.Ensure(remove, install)
	if err == nil {
		l.logger.WithFields(map[string]interface{}{
			"uninstalled":  remove,
			"installed":    install,
			"needs_reboot": reboot,
		}).Debug("Dependency ensured")
	}
	return reboot, errors.Wrapf(err, "uninstalling version %s and installing version %s failed for dependency %s", remove, install, l.unwrapped.String())
}
