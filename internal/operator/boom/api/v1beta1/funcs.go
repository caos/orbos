package v1beta1

import (
	"fmt"

	"github.com/caos/orbos/v5/pkg/tree"
)

func ParseToolset(desiredTree *tree.Tree) (*Toolset, error) {
	desiredKind := &Toolset{}
	if err := desiredTree.Original.Decode(desiredKind); err != nil {
		return nil, fmt.Errorf("parsing desired state failed: %w", err)
	}

	return desiredKind, nil
}
