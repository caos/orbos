package resources

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
)

type AdaptFuncToEnsure func(monitor mntr.Monitor, desired *tree.Tree, current *tree.Tree) (QueryFunc, error)
type AdaptFuncToDelete func(monitor mntr.Monitor, desired *tree.Tree, current *tree.Tree) (DestroyFunc, error)

type EnsureFunc func(*kubernetes.Client) error

type DestroyFunc func(*kubernetes.Client) error

type QueryFunc func(*kubernetes.Client) (EnsureFunc, error)
