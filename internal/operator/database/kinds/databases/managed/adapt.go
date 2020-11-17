package managed

import (
	"strconv"
	"strings"

	core2 "github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/managed/certificate"
	kubernetes2 "github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/secret"

	corev1 "k8s.io/api/core/v1"

	"github.com/caos/orbos/internal/operator/database/kinds/backups"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/core"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/managed/rbac"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/managed/services"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/managed/statefulset"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes/resources/pdb"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
)

func AdaptFunc(
	labels map[string]string,
	namespace string,
	timestamp string,
	nodeselector map[string]string,
	tolerations []corev1.Toleration,
	version string,
	features []string,
) func(
	monitor mntr.Monitor,
	desired *tree.Tree,
	current *tree.Tree,
) (
	core2.QueryFunc,
	core2.DestroyFunc,
	map[string]*secret.Secret,
	error,
) {

	internalLabels := map[string]string{}
	for k, v := range labels {
		internalLabels[k] = v
	}
	internalLabels["app.kubernetes.io/component"] = "iam-database"

	sfsName := "cockroachdb"
	pdbName := sfsName + "-budget"
	serviceAccountName := sfsName
	publicServiceName := sfsName + "-public"
	cockroachPort := int32(26257)
	cockroachHTTPPort := int32(8080)
	image := "cockroachdb/cockroach:v20.1.5"

	return func(
		monitor mntr.Monitor,
		desired *tree.Tree,
		current *tree.Tree,
	) (
		core2.QueryFunc,
		core2.DestroyFunc,
		map[string]*secret.Secret,
		error,
	) {
		internalMonitor := monitor.WithField("kind", "managedDatabase")
		allSecrets := map[string]*secret.Secret{}

		desiredKind, err := parseDesiredV0(desired)
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "parsing desired state failed")
		}
		desired.Parsed = desiredKind

		if !monitor.IsVerbose() && desiredKind.Spec.Verbose {
			internalMonitor.Verbose()
		}
		queryCert, destroyCert, addUser, deleteUser, listUsers, err := certificate.AdaptFunc(internalMonitor, namespace, internalLabels, desiredKind.Spec.ClusterDns)
		if err != nil {
			return nil, nil, nil, err
		}
		addRoot, err := addUser("root")
		if err != nil {
			return nil, nil, nil, err
		}
		destroyRoot, err := deleteUser("root")
		if err != nil {
			return nil, nil, nil, err
		}

		queryRBAC, destroyRBAC, err := rbac.AdaptFunc(internalMonitor, namespace, serviceAccountName, internalLabels)

		querySFS, destroySFS, ensureInit, checkDBReady, listDatabases, err := statefulset.AdaptFunc(
			internalMonitor,
			namespace,
			sfsName,
			image,
			internalLabels,
			serviceAccountName,
			desiredKind.Spec.ReplicaCount,
			desiredKind.Spec.StorageCapacity,
			cockroachPort,
			cockroachHTTPPort,
			desiredKind.Spec.StorageClass,
			desiredKind.Spec.NodeSelector,
			desiredKind.Spec.Tolerations,
			desiredKind.Spec.Resources,
		)
		if err != nil {
			return nil, nil, nil, err
		}

		queryS, destroyS, err := services.AdaptFunc(internalMonitor, namespace, publicServiceName, sfsName, internalLabels, cockroachPort, cockroachHTTPPort)

		//externalName := "cockroachdb-public." + namespaceStr + ".svc.cluster.local"
		//queryES, destroyES, err := service.AdaptFunc("cockroachdb-public", "default", labels, []service.Port{}, "ExternalName", map[string]string{}, false, "", externalName)
		//if err != nil {
		//	return nil, nil, err
		//}

		queryPDB, err := pdb.AdaptFuncToEnsure(namespace, pdbName, internalLabels, "1")
		if err != nil {
			return nil, nil, nil, err
		}

		destroyPDB, err := pdb.AdaptFuncToDestroy(namespace, pdbName)
		if err != nil {
			return nil, nil, nil, err
		}

		currentDB := &Current{
			Common: &tree.Common{
				Kind:    "databases.caos.ch/ManagedDatabase",
				Version: "v0",
			},
			Current: &CurrentDB{
				CA: &certificate.Current{},
			},
		}
		current.Parsed = currentDB

		queriers := make([]core2.QueryFunc, 0)
		for _, feature := range features {
			if feature == "database" {
				queriers = append(queriers,
					queryRBAC,
					queryCert,
					addRoot,
					core2.ResourceQueryToZitadelQuery(querySFS),
					core2.ResourceQueryToZitadelQuery(queryPDB),
					queryS,
					core2.EnsureFuncToQueryFunc(ensureInit),
				)
			}
		}

		featureRestore := false
		destroyers := make([]core2.DestroyFunc, 0)
		for _, feature := range features {
			if feature == "database" {
				destroyers = append(destroyers,
					core2.ResourceDestroyToZitadelDestroy(destroyPDB),
					destroyS,
					core2.ResourceDestroyToZitadelDestroy(destroySFS),
					destroyRBAC,
					destroyCert,
					destroyRoot,
				)
			} else if feature == "restore" {
				featureRestore = true
			}
		}

		if desiredKind.Spec.Backups != nil {

			oneBackup := false
			for backupName := range desiredKind.Spec.Backups {
				if timestamp != "" && strings.HasPrefix(timestamp, backupName) {
					oneBackup = true
				}
			}

			for backupName, desiredBackup := range desiredKind.Spec.Backups {
				currentBackup := &tree.Tree{}
				if timestamp == "" || !oneBackup || (timestamp != "" && strings.HasPrefix(timestamp, backupName)) {
					queryB, destroyB, secrets, err := backups.GetQueryAndDestroyFuncs(
						internalMonitor,
						desiredBackup,
						currentBackup,
						backupName,
						namespace,
						internalLabels,
						checkDBReady,
						strings.TrimPrefix(timestamp, backupName+"."),
						nodeselector,
						tolerations,
						version,
						features,
					)
					if err != nil {
						return nil, nil, nil, err
					}

					secret.AppendSecrets("", allSecrets, secrets)
					destroyers = append(destroyers, destroyB)
					queriers = append(queriers, queryB)
				}
			}
		}

		return func(k8sClient *kubernetes2.Client, queried map[string]interface{}) (core2.EnsureFunc, error) {
				if !featureRestore {
					currentDB.Current.Port = strconv.Itoa(int(cockroachPort))
					currentDB.Current.URL = publicServiceName
					currentDB.Current.ReadyFunc = checkDBReady
					currentDB.Current.AddUserFunc = func(user string) (core2.QueryFunc, error) {
						return addUser(user)
					}
					currentDB.Current.DeleteUserFunc = func(user string) (core2.DestroyFunc, error) {
						return deleteUser(user)
					}
					currentDB.Current.ListUsersFunc = listUsers
					currentDB.Current.ListDatabasesFunc = listDatabases

					core.SetQueriedForDatabase(queried, current)
					internalMonitor.Info("set current state of managed database")
				}

				ensure, err := core2.QueriersToEnsureFunc(internalMonitor, true, queriers, k8sClient, queried)
				return ensure, err
			},
			core2.DestroyersToDestroyFunc(internalMonitor, destroyers),
			allSecrets,
			nil
	}
}
