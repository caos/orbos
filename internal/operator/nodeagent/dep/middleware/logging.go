package middleware

import (
	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/operator/common"
	"github.com/caos/orbiter/internal/operator/nodeagent"
	"github.com/caos/orbiter/mntr"
)

type loggedDep struct {
	monitor mntr.Monitor
	*wrapped
	unwrapped nodeagent.Installer
}

func AddLogging(monitor mntr.Monitor, original nodeagent.Installer) Installer {
	return &loggedDep{
		monitor.WithFields(map[string]interface{}{
			"dependency": original,
		}),
		&wrapped{original},
		Unwrap(original),
	}
}

func (l *loggedDep) Current() (common.Package, error) {
	current, err := l.unwrapped.Current()
	if err == nil {
		l.monitor.WithFields(map[string]interface{}{
			"version": current,
		}).Debug("Queried current dependency version")
	}
	return current, errors.Wrapf(err, "querying installed package for dependency %s failed", l.String())
}

func (l *loggedDep) Ensure(remove common.Package, install common.Package) error {
	return errors.Wrapf(
		l.unwrapped.Ensure(remove, install),
		"uninstalling version %s and installing version %s failed for dependency %s",
		remove,
		install,
		l.unwrapped.String())
}
