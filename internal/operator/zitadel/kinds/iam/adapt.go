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
	"github.com/caos/orbos/internal/operator/zitadel/kinds/networking"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func AdaptFunc(features ...string) zitadel.AdaptFunc {
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

		namespaceStr := "caos-zitadel"
		labels := map[string]string{
			"app.kubernetes.io/managed-by": "zitadel.caos.ch",
			"app.kubernetes.io/part-of":    "zitadel",
		}

		cmName := "zitadel-vars"
		certPath := "$HOME/dbsecrets-zitadel"
		secretName := "zitadel-secret"
		secretPath := "/secret"
		consoleCMName := "console-config"
		secretVarsName := "zitadel-secrets-vars"
		imagePullSecretName := "public-github-packages"

		databaseCurrent := &tree.Tree{}
		queryDB, destroyDB, err := databases.GetQueryAndDestroyFuncs(monitor, desiredKind.Spec.Database, databaseCurrent)
		if err != nil {
			return nil, nil, err
		}

		queryNS, destroyNS, err := namespace.AdaptFunc(namespaceStr)
		if err != nil {
			return nil, nil, err
		}

		queryC, destroyC, err := configuration.AdaptFunc(namespaceStr, labels, desiredKind.Spec.Configuration, cmName, certPath, secretName, secretPath, consoleCMName, secretVarsName)
		if err != nil {
			return nil, nil, err
		}

		queryS, destroyS, err := services.AdaptFunc(namespaceStr, labels)
		if err != nil {
			return nil, nil, err
		}

		queryIPS, destroyIPS, err := imagepullsecret.AdaptFunc(namespaceStr, imagePullSecretName, labels)
		if err != nil {
			return nil, nil, err
		}

		queryD, destroyD, err := deployment.AdaptFunc(namespaceStr, labels, desiredKind.Spec.ReplicaCount, desiredKind.Spec.Version, imagePullSecretName, cmName, certPath, secretName, secretPath, consoleCMName, secretVarsName)
		if err != nil {
			return nil, nil, err
		}

		queryM, destroyM, err := migration.AdaptFunc(namespaceStr, labels)
		if err != nil {
			return nil, nil, err
		}

		networkingCurrent := &tree.Tree{}
		queryNW, destroyNW, err := networking.GetQueryAndDestroyFuncs(monitor, desiredKind.Spec.Networking, networkingCurrent)
		if err != nil {
			return nil, nil, err
		}

		queriers := make([]zitadel.QueryFunc, 0)
		for _, feature := range features {
			switch feature {
			case "networking":
				//networking
				queriers = append(queriers, queryNW)
			case "zitadel":
				queriers = append(queriers, //namespace
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
				)
			}
		}

		destroyers := make([]zitadel.DestroyFunc, 0)
		for _, feature := range features {
			switch feature {
			case "networking":
				destroyers = append(destroyers, destroyNW)
			case "zitadel":
				destroyers = append(destroyers, //namespace
					destroyS,
					destroyM,
					destroyC,
					zitadel.ResourceDestroyToZitadelDestroy(destroyIPS),
					zitadel.ResourceDestroyToZitadelDestroy(destroyD),
					destroyDB,
					zitadel.ResourceDestroyToZitadelDestroy(destroyNS),
				)
			}
		}

		return func(k8sClient *kubernetes.Client, _ map[string]interface{}) (zitadel.EnsureFunc, error) {
				queried := map[string]interface{}{
					"database":   databaseCurrent,
					"networking": networkingCurrent,
				}
				monitor.WithField("queriers", len(queriers)).Info("Querying")
				return zitadel.QueriersToEnsureFunc(queriers, k8sClient, queried)
			},
			func(k8sClient *kubernetes.Client) error {
				monitor.WithField("destroyers", len(destroyers)).Info("Destroying")
				return zitadel.DestroyersToDestroyFunc(destroyers)(k8sClient)
			},
			nil
	}
}
