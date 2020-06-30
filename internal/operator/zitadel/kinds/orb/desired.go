package orb

import (
	"github.com/caos/orbos/internal/tree"
	"github.com/pkg/errors"
)

type DesiredV0 struct {
	Common *tree.Common `yaml:",inline"`
	Spec   struct {
		Verbose bool
	}
	Database *tree.Tree
	IAM      *tree.Tree
}

func ParseDesiredV0(desiredTree *tree.Tree) (*DesiredV0, error) {
	desiredKind := &DesiredV0{Common: desiredTree.Common}

	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}
	desiredKind.Common.Version = "v0"

	return desiredKind, nil
}
