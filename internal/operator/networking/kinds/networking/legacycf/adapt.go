package legacycf

import (
	core2 "github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/internal/operator/networking/kinds/networking/core"
	"github.com/caos/orbos/mntr"
	kubernetes2 "github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
)

func AdaptFunc(
	namespace string,
	labels map[string]string,
) core2.AdaptFunc {
	return func(
		monitor mntr.Monitor,
		desiredTree *tree.Tree,
		currentTree *tree.Tree,
	) (
		core2.QueryFunc,
		core2.DestroyFunc,
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

		internalSpec, current := desiredKind.Spec.Internal(namespace, labels)

		legacyQuerier, legacyDestroyer, readyCertificate, err := adaptFunc(monitor, internalSpec)
		current.ReadyCertificate = readyCertificate

		queriers := []core2.QueryFunc{
			legacyQuerier,
		}
		currentTree.Parsed = current

		return func(k8sClient *kubernetes2.Client, queried map[string]interface{}) (core2.EnsureFunc, error) {
				core.SetQueriedForNetworking(queried, currentTree)
				internalMonitor.Info("set current state legacycf")

				return core2.QueriersToEnsureFunc(internalMonitor, true, queriers, k8sClient, queried)
			},
			core2.DestroyersToDestroyFunc(internalMonitor, []core2.DestroyFunc{legacyDestroyer}),
			nil
	}
}
