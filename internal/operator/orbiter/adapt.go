package orbiter

import (
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
)

type AdaptFunc func(
	monitor mntr.Monitor,
	finishedChan chan struct{},
	desired *tree.Tree,
	current *tree.Tree,
) (
	QueryFunc,
	DestroyFunc,
	ConfigureFunc,
	bool,
	map[string]*secret.Secret,
	error,
)
