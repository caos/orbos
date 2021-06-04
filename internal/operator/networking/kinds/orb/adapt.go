package orb

import (
	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/internal/operator/networking/kinds/networking"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/labels"
	"github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
)

func OperatorSelector() *labels.Selector {
	return labels.OpenOperatorSelector("ORBOS", "networking.caos.ch")
}

func AdaptFunc(ID string, binaryVersion *string, gitops bool) core.AdaptFunc {

	namespaceStr := "caos-zitadel"
	return func(
		monitor mntr.Monitor,
		desiredTree *tree.Tree,
		currentTree *tree.Tree,
	) (queryFunc core.QueryFunc,
		destroyFunc core.DestroyFunc,
		secrets map[string]*secret.Secret,
		existing map[string]*secret.Existing,
		migrate bool,
		err error,
	) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		orbMonitor := monitor.WithField("kind", "orb")

		desiredKind, err := ParseDesiredV0(desiredTree)
		if err != nil {
			return nil, nil, nil, nil, false, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind
		currentTree = &tree.Tree{}

		if desiredKind.Spec.Verbose && !orbMonitor.IsVerbose() {
			orbMonitor = orbMonitor.Verbose()
		}

		operatorLabels := mustDatabaseOperator(binaryVersion)
		networkingCurrent := &tree.Tree{}
		queryNW, destroyNW, secrets, existing, migrate, err := networking.GetQueryAndDestroyFuncs(orbMonitor, ID, operatorLabels, desiredKind.Networking, networkingCurrent, namespaceStr)
		if err != nil {
			return nil, nil, nil, nil, false, err
		}

		queriers := []core.QueryFunc{
			queryNW,
			core.EnsureFuncToQueryFunc(Reconcile(monitor, desiredKind.Spec, gitops)),
		}

		destroyers := []core.DestroyFunc{
			destroyNW,
		}

		currentTree.Parsed = &DesiredV0{
			Common: &tree.Common{
				Kind:    "networking.caos.ch/Orb",
				Version: "v0",
			},
			Networking: networkingCurrent,
		}

		return func(k8sClient kubernetes.ClientInt, _ map[string]interface{}) (core.EnsureFunc, error) {
				queried := map[string]interface{}{}
				monitor.WithField("queriers", len(queriers)).Info("Querying")
				return core.QueriersToEnsureFunc(monitor, true, queriers, k8sClient, queried)
			},
			func(k8sClient kubernetes.ClientInt) error {
				monitor.WithField("destroyers", len(queriers)).Info("Destroy")
				return core.DestroyersToDestroyFunc(monitor, destroyers)(k8sClient)
			},
			secrets,
			existing,
			migrate,
			nil
	}
}
