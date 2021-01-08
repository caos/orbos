package orb

import (
	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/internal/operator/database/kinds/databases"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources/namespace"
	"github.com/caos/orbos/pkg/labels"
	"github.com/caos/orbos/pkg/secret"
	"github.com/caos/orbos/pkg/tree"
	"github.com/caos/orbos/pkg/treelabels"
	"github.com/pkg/errors"
)

func OperatorSelector() *labels.Selector {
	return labels.OpenOperatorSelector("database")
}

func AdaptFunc(timestamp string, binaryVersion *string, features ...string) core.AdaptFunc {
	namespaceStr := "caos-zitadel"

	return func(monitor mntr.Monitor, orbDesiredTree *tree.Tree, currentTree *tree.Tree) (queryFunc core.QueryFunc, destroyFunc core.DestroyFunc, secrets map[string]*secret.Secret, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", orbDesiredTree.Common.Kind)
		}()

		orbMonitor := monitor.WithField("kind", "orb")

		desiredKind, err := parseDesiredV0(orbDesiredTree)
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "parsing desired state failed")
		}
		orbDesiredTree.Parsed = desiredKind
		currentTree = &tree.Tree{}

		if desiredKind.Spec.Verbose && !orbMonitor.IsVerbose() {
			orbMonitor = orbMonitor.Verbose()
		}

		queryNS, err := namespace.AdaptFuncToEnsure(namespaceStr)
		if err != nil {
			return nil, nil, nil, err
		}
		destroyNS, err := namespace.AdaptFuncToDestroy(namespaceStr)
		if err != nil {
			return nil, nil, nil, err
		}

		databaseCurrent := &tree.Tree{}

		operatorLabels := mustDatabaseOperator(binaryVersion)

		queryDB, destroyDB, secrets, err := databases.GetQueryAndDestroyFuncs(
			orbMonitor,
			desiredKind.Database,
			databaseCurrent,
			namespaceStr,
			treelabels.MustForAPI(desiredKind.Database, operatorLabels),
			timestamp,
			desiredKind.Spec.NodeSelector,
			desiredKind.Spec.Tolerations,
			desiredKind.Spec.Version,
			features,
		)

		if err != nil {
			return nil, nil, nil, err
		}
		queriers := []core.QueryFunc{
			core.ResourceQueryToZitadelQuery(queryNS),
			queryDB,
		}
		if desiredKind.Spec.SelfReconciling {
			queriers = append(queriers,
				core.EnsureFuncToQueryFunc(Reconcile(monitor, orbDesiredTree)),
			)
		}

		destroyers := []core.DestroyFunc{
			core.ResourceDestroyToZitadelDestroy(destroyNS),
			destroyDB,
		}

		currentTree.Parsed = &DesiredV0{
			Common: &tree.Common{
				Kind:    "databases.caos.ch/Orb",
				Version: "v0",
			},
			Database: databaseCurrent,
		}

		return func(k8sClient kubernetes.ClientInt, queried map[string]interface{}) (core.EnsureFunc, error) {
				if queried == nil {
					queried = map[string]interface{}{}
				}
				monitor.WithField("queriers", len(queriers)).Info("Querying")
				return core.QueriersToEnsureFunc(monitor, true, queriers, k8sClient, queried)
			},
			func(k8sClient kubernetes.ClientInt) error {
				monitor.WithField("destroyers", len(queriers)).Info("Destroy")
				return core.DestroyersToDestroyFunc(monitor, destroyers)(k8sClient)
			},
			secrets,
			nil
	}
}
