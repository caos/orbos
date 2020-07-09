package networking

import (
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/networking/legacycf"
	"github.com/caos/orbos/internal/secret"
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
	case "zitadel.caos.ch/LegacyCloudflare":
		return legacycf.AdaptFunc()(monitor, desiredTree, currentTree)
	default:
		return nil, nil, errors.Errorf("unknown networking kind %s", desiredTree.Common.Kind)
	}
}

func GetSecrets(
	monitor mntr.Monitor,
	networkingTree *tree.Tree,
) (
	map[string]*secret.Secret,
	error,
) {
	switch networkingTree.Common.Kind {
	case "zitadel.caos.ch/LegacyCloudflare":
		return legacycf.SecretsFunc()(
			monitor,
			networkingTree,
		)
	default:
		return nil, errors.Errorf("unknown networking kind %s", networkingTree.Common.Kind)
	}
}
