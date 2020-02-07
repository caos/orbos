package middleware

import (
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/nodeagent"
	"github.com/caos/orbiter/logging"
)

type loggedDep struct {
	logger logging.Logger
	*wrapped
	unwrapped nodeagent.Installer
}

func AddLogging(logger logging.Logger, original nodeagent.Installer) Installer {
	return &loggedDep{
		logger.WithFields(map[string]interface{}{
			"dependency": original,
		}),
		&wrapped{original},
		Unwrap(original),
	}
}

func (l *loggedDep) Current() (common.Package, error) {
	current, err := l.unwrapped.Current()
	if err == nil {
		l.logger.WithFields(map[string]interface{}{
			"version": current,
		}).Debug("Queried current dependency version")
	}
	return current, errors.Wrapf(err, "querying installed package for dependency %s failed", l.String())
}

func (l *loggedDep) Ensure(remove common.Package, install common.Package) error {
	err := l.unwrapped.Ensure(remove, install)
	if err == nil {
		l.logger.WithFields(map[string]interface{}{
			"uninstalled": remove,
			"installed":   install,
		}).Debug("Dependency ensured")
	}
	return errors.Wrapf(err, "uninstalling version %s and installing version %s failed for dependency %s", remove, install, l.unwrapped.String())
}
