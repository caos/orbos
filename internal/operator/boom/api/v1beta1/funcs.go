package v1beta1

import (
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
)

func ParseToolset(desiredTree *tree.Tree) (*Toolset, error) {
	desiredKind := &Toolset{}
	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, errors.Wrap(err, "parsing desired state failed")
	}

	return desiredKind, nil
}
