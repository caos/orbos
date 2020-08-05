package zitadel

import (
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func BackupListFunc() func(monitor mntr.Monitor, desired *tree.Tree) (strings []string, err error) {
	return func(monitor mntr.Monitor, desired *tree.Tree) (strings []string, err error) {
		desiredKind, err := parseDesiredV0(desired)
		if err != nil {
			return nil, errors.Wrap(err, "parsing desired state failed")
		}
		desired.Parsed = desiredKind

		if !monitor.IsVerbose() && desiredKind.Spec.Verbose {
			monitor.Verbose()
		}

		return databases.GetBackupList(monitor, desiredKind.Database)
	}
}
