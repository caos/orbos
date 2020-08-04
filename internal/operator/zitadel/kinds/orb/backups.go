package orb

import (
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func BackupListFunc() func(monitor mntr.Monitor, desiredTree *tree.Tree) (strings []string, err error) {
	return func(monitor mntr.Monitor, desiredTree *tree.Tree) (strings []string, err error) {
		desiredKind, err := ParseDesiredV0(desiredTree)
		if err != nil {
			return nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		if desiredKind.Spec.Verbose && !monitor.IsVerbose() {
			monitor = monitor.Verbose()
		}

		return iam.BackupListFunc()(monitor, desiredKind.IAM)
	}
}
