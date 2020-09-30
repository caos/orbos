package orb

import (
	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/internal/operator/database/kinds/databases"
	"github.com/caos/orbos/mntr"
	kubernetes2 "github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
)

func AdaptFunc(timestamp string, features ...string) core.AdaptFunc {
	namespaceStr := "caos-zitadel"
	labels := map[string]string{
		"app.kubernetes.io/managed-by": "database.caos.ch",
		"app.kubernetes.io/part-of":    "database",
	}

	return func(monitor mntr.Monitor, desiredTree *tree.Tree, currentTree *tree.Tree) (queryFunc core.QueryFunc, destroyFunc core.DestroyFunc, err error) {
		defer func() {
			err = errors.Wrapf(err, "building %s failed", desiredTree.Common.Kind)
		}()

		orbMonitor := monitor.WithField("kind", "orb")

		desiredKind, err := parseDesiredV0(desiredTree)
		if err != nil {
			return nil, nil, errors.Wrap(err, "parsing desired state failed")
		}
		desiredTree.Parsed = desiredKind
		currentTree = &tree.Tree{}

		if desiredKind.Spec.Verbose && !orbMonitor.IsVerbose() {
			orbMonitor = orbMonitor.Verbose()
		}

		databaseCurrent := &tree.Tree{}
		queryDB, destroyDB, err := databases.GetQueryAndDestroyFuncs(
			orbMonitor,
			desiredKind.Database,
			databaseCurrent,
			namespaceStr,
			labels,
			timestamp,
			desiredKind.Spec.NodeSelector,
			desiredKind.Spec.Tolerations,
			features,
		)

		if err != nil {
			return nil, nil, err
		}
		queriers := []core.QueryFunc{
			queryDB,
		}

		destroyers := []core.DestroyFunc{
			destroyDB,
		}

		currentTree.Parsed = &DesiredV0{
			Common: &tree.Common{
				Kind:    "zitadel.caos.ch/Orb",
				Version: "v0",
			},
			Database: databaseCurrent,
		}

		return func(k8sClient *kubernetes2.Client, queried map[string]interface{}) (core.EnsureFunc, error) {
				if queried == nil {
					queried = map[string]interface{}{}
				}
				monitor.WithField("queriers", len(queriers)).Info("Querying")
				return core.QueriersToEnsureFunc(monitor, true, queriers, k8sClient, queried)
			},
			func(k8sClient *kubernetes2.Client) error {
				monitor.WithField("destroyers", len(queriers)).Info("Destroy")
				return core.DestroyersToDestroyFunc(monitor, destroyers)(k8sClient)
			},
			nil
	}
}
