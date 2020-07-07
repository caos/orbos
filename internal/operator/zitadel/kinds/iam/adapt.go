package iam

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/namespace"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/configuration"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/deployment"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/imagepullsecret"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/migration"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/services"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func AdaptFunc() zitadel.AdaptFunc {
	return func(
		monitor mntr.Monitor,
		desired *tree.Tree,
		current *tree.Tree,
	) (
		zitadel.QueryFunc,
		zitadel.DestroyFunc,
		error,
	) {
		desiredKind, err := parseDesiredV0(desired)
		if err != nil {
			return nil, nil, errors.Wrap(err, "parsing desired state failed")
		}
		desired.Parsed = desiredKind

		databaseCurrent := &tree.Tree{}
		queryDB, destroyDB, err := databases.GetQueryAndDestroyFuncs(monitor, desiredKind.Spec.Database, databaseCurrent)
		if err != nil {
			return nil, nil, err
		}

		namespaceStr := "caos-zitadel"
		labels := map[string]string{"app.kubernetes.io/managed-by": "zitadel.caos.ch"}

		queryNS, destroyNS, err := namespace.AdaptFunc(namespaceStr)
		if err != nil {
			return nil, nil, err
		}

		queryC, destroyC, err := configuration.AdaptFunc(namespaceStr, labels, desiredKind.Spec.Configuration)
		if err != nil {
			return nil, nil, err
		}

		queryS, destroyS, err := services.AdaptFunc(namespaceStr, labels)
		if err != nil {
			return nil, nil, err
		}

		queryIPS, destroyIPS, err := imagepullsecret.AdaptFunc(namespaceStr, labels)
		if err != nil {
			return nil, nil, err
		}

		queryD, destroyD, err := deployment.AdaptFunc(namespaceStr, labels, desiredKind.Spec.ReplicaCount, desiredKind.Spec.Version)
		if err != nil {
			return nil, nil, err
		}

		queryM, destroyM, err := migration.AdaptFunc(namespaceStr, labels)
		if err != nil {
			return nil, nil, err
		}

		queriers := make([]zitadel.QueryFunc, 0)
		queriers = []zitadel.QueryFunc{
			//namespace
			zitadel.ResourceQueryToZitadelQuery(queryNS),
			//database
			queryDB,
			//migration
			queryM,
			//services
			queryS,
			//configuration
			queryC,
			zitadel.ResourceQueryToZitadelQuery(queryIPS),
			zitadel.ResourceQueryToZitadelQuery(queryD),
		}

		destroyers := make([]zitadel.DestroyFunc, 0)
		destroyers = []zitadel.DestroyFunc{
			destroyS,
			destroyM,
			destroyC,
			zitadel.ResourceDestroyToZitadelDestroy(destroyIPS),
			zitadel.ResourceDestroyToZitadelDestroy(destroyD),
			destroyDB,
			zitadel.ResourceDestroyToZitadelDestroy(destroyNS),
		}

		return func(k8sClient *kubernetes.Client, _ map[string]interface{}) (zitadel.EnsureFunc, error) {
				queried := map[string]interface{}{"database": databaseCurrent}

				return zitadel.QueriersToEnsureFunc(queriers, k8sClient, queried)
			},
			zitadel.DestroyersToDestroyFunc(destroyers),
			nil
	}
}
