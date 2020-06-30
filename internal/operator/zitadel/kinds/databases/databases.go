package databases

import (
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/managed"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/provided"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func GetQueryAndDestroyFuncs(
	monitor mntr.Monitor,
	desiredTree *tree.Tree,
	currentTree *tree.Tree,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {
	switch desiredTree.Common.Kind {
	case "zitadel.caos.ch/ManagedDatabase":
		return managed.AdaptFunc()(monitor, desiredTree, currentTree)
	case "zitadel.caos.ch/ProvidedDatabse":
		return provided.AdaptFunc()(monitor, desiredTree, currentTree)
	default:
		return nil, nil, errors.Errorf("unknown provider kind %s", desiredTree.Common.Kind)
	}
}
