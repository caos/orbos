package managed

import (
	"strconv"
	"strings"

	"github.com/caos/orbos/pkg/labels"

	core2 "github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/managed/certificate"
	"github.com/caos/orbos/pkg/secret"

	corev1 "k8s.io/api/core/v1"

	"github.com/caos/orbos/internal/operator/database/kinds/backups"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/core"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/managed/rbac"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/managed/services"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/managed/statefulset"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources/pdb"
	"github.com/caos/orbos/pkg/tree"
	"github.com/pkg/errors"
)

const (
	component          = "database"
	sfsName            = "cockroachdb"
	pdbName            = sfsName + "-budget"
	serviceAccountName = sfsName
	publicServiceName  = sfsName + "-public"
	privateServiceName = sfsName
	cockroachPort      = int32(26257)
	cockroachHTTPPort  = int32(8080)
	image              = "cockroachdb/cockroach:v20.2.3"
)

func PublicServiceNameSelector() *labels.Selector {
	return labels.OpenNameSelector(component, publicServiceName)
}

func AdaptFunc(
	operatorLabels *labels.Operator,
	apiLabels *labels.API,
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
		componentLabels := labels.MustForComponent(apiLabels, "cockroachdb")
		internalMonitor := monitor.WithField("component", "cockroachdb")
		allSecrets := map[string]*secret.Secret{}

		desiredKind, err := parseDesiredV0(desired)
		if err != nil {
			return nil, nil, nil, errors.Wrap(err, "parsing desired state failed")
		}
		desired.Parsed = desiredKind

		if !monitor.IsVerbose() && desiredKind.Spec.Verbose {
			internalMonitor.Verbose()
		}

		var (
			isFeatureDatabase bool
			isFeatureRestore  bool
		)
		for _, feature := range features {
			switch feature {
			case "database":
				isFeatureDatabase = true
			case "restore":
				isFeatureRestore = true
			}
		}

		queryCert, destroyCert, addUser, deleteUser, listUsers, err := certificate.AdaptFunc(internalMonitor, namespace, componentLabels, desiredKind.Spec.ClusterDns, isFeatureDatabase)
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

		queryRBAC, destroyRBAC, err := rbac.AdaptFunc(internalMonitor, namespace, labels.MustForName(componentLabels, serviceAccountName))

		cockroachNameLabels := labels.MustForName(componentLabels, sfsName)
		cockroachSelector := labels.DeriveNameSelector(cockroachNameLabels, false)
		cockroachSelectabel := labels.AsSelectable(cockroachNameLabels)
		querySFS, destroySFS, ensureInit, checkDBReady, listDatabases, err := statefulset.AdaptFunc(
			internalMonitor,
			cockroachSelectabel,
			cockroachSelector,
			desiredKind.Spec.Force,
			namespace,
			image,
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

		queryS, destroyS, err := services.AdaptFunc(
			internalMonitor,
			namespace,
			labels.MustForName(componentLabels, publicServiceName),
			labels.MustForName(componentLabels, privateServiceName),
			cockroachSelector,
			cockroachPort,
			cockroachHTTPPort,
		)

		//externalName := "cockroachdb-public." + namespaceStr + ".svc.cluster.local"
		//queryES, destroyES, err := service.AdaptFunc("cockroachdb-public", "default", labels, []service.Port{}, "ExternalName", map[string]string{}, false, "", externalName)
		//if err != nil {
		//	return nil, nil, err
		//}

		queryPDB, err := pdb.AdaptFuncToEnsure(namespace, labels.MustForName(componentLabels, pdbName), cockroachSelector, "1")
		if err != nil {
			return nil, nil, nil, err
		}

		destroyPDB, err := pdb.AdaptFuncToDestroy(namespace, pdbName)
		if err != nil {
			return nil, nil, nil, err
		}

		currentDB := &Current{
			Common: &tree.Common{
				Kind:    "databases.caos.ch/CockroachDB",
				Version: "v0",
			},
			Current: &CurrentDB{
				CA: &certificate.Current{},
			},
		}
		current.Parsed = currentDB

		queriers := make([]core2.QueryFunc, 0)
		if isFeatureDatabase {
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

		destroyers := make([]core2.DestroyFunc, 0)
		if isFeatureDatabase {
			destroyers = append(destroyers,
				core2.ResourceDestroyToZitadelDestroy(destroyPDB),
				destroyS,
				core2.ResourceDestroyToZitadelDestroy(destroySFS),
				destroyRBAC,
				destroyCert,
				destroyRoot,
			)
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
						operatorLabels,
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

					secret.AppendSecrets(backupName, allSecrets, secrets)
					destroyers = append(destroyers, destroyB)
					queriers = append(queriers, queryB)
				}
			}
		}

		return func(k8sClient kubernetes.ClientInt, queried map[string]interface{}) (core2.EnsureFunc, error) {
				if !isFeatureRestore {
					queriedCurrentDB, err := core.ParseQueriedForDatabase(queried)
					if err != nil || queriedCurrentDB == nil {
						// TODO: query system state
						currentDB.Current.Port = strconv.Itoa(int(cockroachPort))
						currentDB.Current.URL = publicServiceName
						currentDB.Current.ReadyFunc = checkDBReady
						currentDB.Current.AddUserFunc = addUser
						currentDB.Current.DeleteUserFunc = deleteUser
						currentDB.Current.ListUsersFunc = listUsers
						currentDB.Current.ListDatabasesFunc = listDatabases

						core.SetQueriedForDatabase(queried, current)
						internalMonitor.Info("set current state of managed database")
					}
				}

				ensure, err := core2.QueriersToEnsureFunc(internalMonitor, true, queriers, k8sClient, queried)
				return ensure, err
			},
			core2.DestroyersToDestroyFunc(internalMonitor, destroyers),
			allSecrets,
			nil
	}
}
