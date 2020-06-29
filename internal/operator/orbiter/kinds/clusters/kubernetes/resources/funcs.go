package resources

import (
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
)

type AdaptFunc func(monitor mntr.Monitor, desired *tree.Tree, current *tree.Tree) (QueryFunc, DestroyFunc, error)

type EnsureFunc func() error

type DestroyFunc func() error

type QueryFunc func() (EnsureFunc, error)
