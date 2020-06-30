package resources

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
)

type AdaptFunc func(monitor mntr.Monitor, desired *tree.Tree, current *tree.Tree) (QueryFunc, DestroyFunc, error)

type EnsureFunc func(client *kubernetes.Client) error

type DestroyFunc func(client *kubernetes.Client) error

type QueryFunc func() (EnsureFunc, error)
