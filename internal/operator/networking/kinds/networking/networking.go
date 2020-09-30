package networking

import (
	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/internal/operator/networking/kinds/networking/legacycf"
	"github.com/caos/orbos/mntr"
	secret2 "github.com/caos/orbos/pkg/secret"
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
	error,
) {
	switch desiredTree.Common.Kind {
	case "zitadel.caos.ch/LegacyCloudflare":
		return legacycf.AdaptFunc(namespace, labels)(monitor, desiredTree, currentTree)
	default:
		return nil, nil, errors.Errorf("unknown networking kind %s", desiredTree.Common.Kind)
	}
}

func GetSecrets(
	monitor mntr.Monitor,
	networkingTree *tree.Tree,
) (
	map[string]*secret2.Secret,
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
