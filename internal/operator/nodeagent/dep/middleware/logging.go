package middleware

import (
	"fmt"
	"github.com/caos/orbos/internal/operator/common"
	"github.com/caos/orbos/internal/operator/nodeagent"
	"github.com/caos/orbos/mntr"
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

func (l *loggedDep) InstalledFilter() []string { return l.unwrapped.InstalledFilter() }

func (l *loggedDep) Current() (common.Package, error) {
	current, err := l.unwrapped.Current()
	if err != nil {
		return current, fmt.Errorf("querying installed package for dependency %s failed: %w", l.String(), err)
	}
	l.monitor.WithFields(map[string]interface{}{
		"version": current,
	}).Debug("Queried current dependency version")

	return current, nil
}

func (l *loggedDep) Ensure(remove common.Package, install common.Package, leaveOSRepositories bool) error {

	var leavingOSREpositories string
	if leaveOSRepositories {
		leavingOSREpositories = "leaving OS repositories "
	}

	if err := l.unwrapped.Ensure(remove, install, leaveOSRepositories); err != nil {
		return fmt.Errorf("uninstalling version %s and installing version %s %sfailed for dependency %s: %w",
			remove,
			install,
			leavingOSREpositories,
			l.unwrapped.String(),
			err)
	}
	return nil
}
