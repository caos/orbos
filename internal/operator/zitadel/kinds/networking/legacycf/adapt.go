package legacycf

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/namespace"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func AdaptFunc() zitadel.AdaptFunc {
	return func(
		monitor mntr.Monitor,
		desiredTree *tree.Tree,
		currentTree *tree.Tree,
	) (
		zitadel.QueryFunc,
		zitadel.DestroyFunc,
		error,
	) {
		desiredKind, err := parseDesired(desiredTree)
		if err != nil {
			return nil, nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		if desiredKind.Spec == nil {
			return nil, nil, errors.New("No specs found")
		}

		if err := desiredKind.Spec.Validate(); err != nil {
			return nil, nil, err
		}
		namespaceStr := "caos-zitadel"

		queryNS, _, err := namespace.AdaptFunc(namespaceStr)
		if err != nil {
			return nil, nil, err
		}

		internalSpec, current := desiredKind.Spec.Internal()
		internalSpec.Namespace = namespaceStr

		legacyQuerier, legacyDestroyer, err := adaptFunc(internalSpec)
		currentTree.Parsed = current

		queriers := []zitadel.QueryFunc{
			zitadel.ResourceQueryToZitadelQuery(queryNS),
			legacyQuerier,
		}

		return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
				return zitadel.QueriersToEnsureFunc(queriers, k8sClient, queried)
			},
			zitadel.DestroyersToDestroyFunc([]zitadel.DestroyFunc{legacyDestroyer}),
			nil
	}
}
