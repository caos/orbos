package legacycf

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func AdaptFunc(
	namespace string,
	originCASecretName string,
	labels map[string]string,
) zitadel.AdaptFunc {
	return func(
		monitor mntr.Monitor,
		desiredTree *tree.Tree,
		currentTree *tree.Tree,
	) (
		zitadel.QueryFunc,
		zitadel.DestroyFunc,
		error,
	) {
		internalMonitor := monitor.WithField("kind", "legacycf")

		desiredKind, err := parseDesired(desiredTree)
		if err != nil {
			return nil, nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind

		if !monitor.IsVerbose() && desiredKind.Spec.Verbose {
			internalMonitor.Verbose()
		}

		if desiredKind.Spec == nil {
			return nil, nil, errors.New("No specs found")
		}

		if err := desiredKind.Spec.Validate(); err != nil {
			return nil, nil, err
		}

		internalSpec, current := desiredKind.Spec.Internal(namespace, originCASecretName, labels)

		legacyQuerier, legacyDestroyer, readyCertificate, err := adaptFunc(monitor, internalSpec)
		current.ReadyCertificate = readyCertificate

		queriers := []zitadel.QueryFunc{
			legacyQuerier,
		}

		return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
				currentTree.Parsed = current
				internalMonitor.Info("set current state legacycf")

				return zitadel.QueriersToEnsureFunc(internalMonitor, true, queriers, k8sClient, queried)
			},
			zitadel.DestroyersToDestroyFunc(internalMonitor, []zitadel.DestroyFunc{legacyDestroyer}),
			nil
	}
}
