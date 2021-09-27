package orbiter

import (
	"github.com/caos/orbos/v5/mntr"
	"github.com/caos/orbos/v5/pkg/secret"
	"github.com/caos/orbos/v5/pkg/tree"
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
