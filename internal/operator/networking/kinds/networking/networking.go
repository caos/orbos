package networking

import (
	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/internal/operator/networking/kinds/networking/legacycf"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/labels"
	"github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
)

func GetQueryAndDestroyFuncs(
	monitor mntr.Monitor,
	desiredTree *tree.Tree,
	currentTree *tree.Tree,
	namespace string,
	operatorLabels *labels.Operator,
) (
	query core.QueryFunc,
	destroy core.DestroyFunc,
	secrets map[string]*secret.Secret,
	apiLabels *labels.API,
	err error,
) {
	switch desiredTree.Common.Kind {
	case "networking.caos.ch/LegacyCloudflare":
		apiLabels = labels.MustForAPI(operatorLabels, "LegacyCloudflare", desiredTree.Common.Version)
		query, destroy, secrets, err = legacycf.AdaptFunc(namespace, apiLabels)(monitor, desiredTree, currentTree)
	default:
		err = errors.Errorf("unknown networking kind %s", desiredTree.Common.Kind)
	}
	return query, destroy, secrets, apiLabels, err
}
