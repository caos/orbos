package iam

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/namespace"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases"
	coredb "github.com/caos/orbos/internal/operator/zitadel/kinds/databases/core"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/ambassador"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/configuration"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/deployment"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/imagepullsecret"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/migration"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/services"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/networking"
	corenw "github.com/caos/orbos/internal/operator/zitadel/kinds/networking/core"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	"sort"
	"strconv"
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
		internalLabels := map[string]string{}
		for k, v := range labels {
			internalLabels[k] = v
		}
		internalLabels["app.kubernetes.io/component"] = "iam"

		cmName := "zitadel-vars"
		certPath := "/home/zitadel/dbsecrets-zitadel"
		secretName := "zitadel-secret"
		secretPath := "/secret"
		consoleCMName := "console-config"
		secretVarsName := "zitadel-secrets-vars"
		secretPasswordName := "zitadel-passwords"
		imagePullSecretName := "public-github-packages"
		grpcServiceName := "grpc-v1"
		grpcPort := 80
		httpServiceName := "http-v1"
		httpPort := 80
		uiServiceName := "ui-v1"
		uiPort := 80
		httpURL := "http://" + httpServiceName + "." + namespaceStr + ":" + strconv.Itoa(httpPort)
		grpcURL := grpcServiceName + "." + namespaceStr + ":" + strconv.Itoa(grpcPort)
		uiURL := "http://" + uiServiceName + "." + namespaceStr
		originCASecretName := "tls-cert-wildcard"

		users, migrationUser := getUsers(desiredKind)

		allZitadelUsers := make([]string, 0)
		for k := range users {
			if k != migrationUser {
				allZitadelUsers = append(allZitadelUsers, k)
			}
		}
		sort.Slice(allZitadelUsers, func(i, j int) bool {
			return allZitadelUsers[i] < allZitadelUsers[j]
		})

		allUsers := make([]string, 0)
		for k := range users {
			allUsers = append(allUsers, k)
		}
		sort.Slice(allUsers, func(i, j int) bool {
			return allUsers[i] < allUsers[j]
		})

		databaseCurrent := &tree.Tree{}
		queryDB, destroyDB, err := databases.GetQueryAndDestroyFuncs(monitor, desiredKind.Database, databaseCurrent, namespaceStr, allUsers, labels)
		if err != nil {
			return nil, nil, err
		}

		queryNS, err := namespace.AdaptFuncToEnsure(namespaceStr)
		if err != nil {
			return nil, nil, err
		}
		destroyNS, err := namespace.AdaptFuncToDestroy(namespaceStr)
		if err != nil {
			return nil, nil, err
		}

		queryC, destroyC, err := configuration.AdaptFunc(
			namespaceStr,
			labels,
			desiredKind.Spec.Configuration,
			cmName,
			certPath,
			secretName,
			secretPath,
			consoleCMName,
			secretVarsName,
			secretPasswordName,
			users,
		)
		if err != nil {
			return nil, nil, err
		}

		queryS, destroyS, err := services.AdaptFunc(namespaceStr, internalLabels, grpcServiceName, grpcPort, httpServiceName, httpPort, uiServiceName, uiPort)
		if err != nil {
			return nil, nil, err
		}

		queryIPS, destroyIPS, err := imagepullsecret.AdaptFunc(namespaceStr, imagePullSecretName, internalLabels)
		if err != nil {
			return nil, nil, err
		}

		queryD, destroyD, err := deployment.AdaptFunc(
			namespaceStr,
			internalLabels,
			desiredKind.Spec.ReplicaCount,
			desiredKind.Spec.Version,
			imagePullSecretName,
			cmName,
			certPath,
			secretName,
			secretPath,
			consoleCMName,
			secretVarsName,
			secretPasswordName,
			allZitadelUsers,
			desiredKind.Spec.NodeSelector,
		)
		if err != nil {
			return nil, nil, err
		}

		queryM, destroyM, err := migration.AdaptFunc(namespaceStr, internalLabels, secretPasswordName, migrationUser, allZitadelUsers)
		if err != nil {
			return nil, nil, err
		}

		networkingCurrent := &tree.Tree{}
		queryNW, destroyNW, err := networking.GetQueryAndDestroyFuncs(monitor, desiredKind.Networking, networkingCurrent, namespaceStr, originCASecretName, labels)
		if err != nil {
			return nil, nil, err
		}

		queryAmbassador, destroyAmbassador, err := ambassador.AdaptFunc(namespaceStr, labels, grpcURL, httpURL, uiURL, originCASecretName)
		if err != nil {
			return nil, nil, err
		}

		queriers := make([]zitadel.QueryFunc, 0)
		for _, feature := range features {
			switch feature {
			case "networking":
				//networking
				queriers = append(queriers,
					zitadel.ResourceQueryToZitadelQuery(queryNS),
					queryNW,
					queryAmbassador,
				)
			case "zitadel":
				queriers = append(queriers, //namespace
					zitadel.ResourceQueryToZitadelQuery(queryNS),
					//database
					queryDB,
					//configuration
					queryC,
					//migration
					queryM,
					//services
					queryS,
					zitadel.ResourceQueryToZitadelQuery(queryIPS),
					zitadel.ResourceQueryToZitadelQuery(queryD),
				)
			}
		}

		destroyers := make([]zitadel.DestroyFunc, 0)
		for _, feature := range features {
			switch feature {
			case "networking":
				destroyers = append(destroyers,
					destroyNW,
					destroyAmbassador,
					zitadel.ResourceDestroyToZitadelDestroy(destroyNS),
				)
			case "zitadel":
				destroyers = append(destroyers, //namespace
					destroyS,
					destroyM,
					zitadel.ResourceDestroyToZitadelDestroy(destroyIPS),
					zitadel.ResourceDestroyToZitadelDestroy(destroyD),
					destroyDB,
					destroyC,
					zitadel.ResourceDestroyToZitadelDestroy(destroyNS),
				)
			}
		}

		return func(k8sClient *kubernetes.Client, _ map[string]interface{}) (zitadel.EnsureFunc, error) {
				queried := map[string]interface{}{}
				corenw.SetQueriedForNetworking(queried, networkingCurrent)
				coredb.SetQueriedForDatabase(queried, databaseCurrent)

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

func getUsers(desired *DesiredV0) (map[string]string, string) {
	passwords := &configuration.Passwords{}
	if desired.Spec != nil && desired.Spec.Configuration != nil && desired.Spec.Configuration.Passwords != nil {
		passwords = desired.Spec.Configuration.Passwords
	}
	users := make(map[string]string, 0)

	migrationUser := "flyway"
	migrationPassword := migrationUser
	if passwords.Migration != nil {
		migrationPassword = passwords.Migration.Value
	}
	users[migrationUser] = migrationPassword

	mgmtUser := "management"
	mgmtPassword := mgmtUser
	if passwords != nil && passwords.Management != nil {
		mgmtPassword = passwords.Management.Value
	}
	users[mgmtUser] = mgmtPassword

	adminUser := "adminapi"
	adminPassword := adminUser
	if passwords != nil && passwords.Adminapi != nil {
		adminPassword = passwords.Adminapi.Value
	}
	users[adminUser] = adminPassword

	authUser := "auth"
	authPassword := authUser
	if passwords != nil && passwords.Auth != nil {
		authPassword = passwords.Auth.Value
	}
	users[authUser] = authPassword

	authzUser := "authz"
	authzPassword := authzUser
	if passwords != nil && passwords.Authz != nil {
		authzPassword = passwords.Authz.Value
	}
	users[authzUser] = authzPassword

	notUser := "notification"
	notPassword := notUser
	if passwords != nil && passwords.Notification != nil {
		notPassword = passwords.Notification.Value
	}
	users[notUser] = notPassword

	return users, migrationUser
}
