package networking

import (
	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/internal/operator/networking/kinds/networking/legacycf"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
)

func GetQueryAndDestroyFuncs(
	monitor mntr.Monitor,
	desiredTree *tree.Tree,
	currentTree *tree.Tree,
	namespace string,
	labels map[string]string,
) (
	core.QueryFunc,
	core.DestroyFunc,
	map[string]*secret.Secret,
	error,
) {
	switch desiredTree.Common.Kind {
	case "zitadel.caos.ch/LegacyCloudflare":
		return legacycf.AdaptFunc(namespace, labels)(monitor, desiredTree, currentTree)
	default:
		return nil, nil, nil, errors.Errorf("unknown networking kind %s", desiredTree.Common.Kind)
	}
}
