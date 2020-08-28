package managed

import (
	"strconv"
	"strings"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/pdb"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/backups"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/core"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/managed/certificate"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/managed/rbac"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/managed/services"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/managed/statefulset"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
)

func AdaptFunc(
	labels map[string]string,
	users []string,
	namespace string,
	timestamp string,
	secretPasswordName string,
	migrationUser string,
	features []string,
) zitadel.AdaptFunc {
	return func(
		monitor mntr.Monitor,
		desired *tree.Tree,
		current *tree.Tree,
	) (
		zitadel.QueryFunc,
		zitadel.DestroyFunc,
		error,
	) {
		internalMonitor := monitor.WithField("kind", "managedDatabase")

		desiredKind, err := parseDesiredV0(desired)
		if err != nil {
			return nil, nil, errors.Wrap(err, "parsing desired state failed")
		}
		desired.Parsed = desiredKind

		if !monitor.IsVerbose() && desiredKind.Spec.Verbose {
			internalMonitor.Verbose()
		}

		interalLabels := map[string]string{}
		for k, v := range labels {
			interalLabels[k] = v
		}
		interalLabels["app.kubernetes.io/component"] = "iam-database"

		sfsName := "cockroachdb"
		pdbName := sfsName + "-budget"
		serviceAccountName := sfsName
		publicServiceName := sfsName + "-public"
		cockroachPort := int32(26257)
		cockroachHTTPPort := int32(8080)
		image := "cockroachdb/cockroach:v20.1.4"

		userList := []string{"root"}
		userList = append(userList, users...)

		queryCert, destroyCert, err := certificate.AdaptFunc(internalMonitor, namespace, userList, interalLabels, desiredKind.Spec.ClusterDns)
		if err != nil {
			return nil, nil, err
		}

		queryRBAC, destroyRBAC, err := rbac.AdaptFunc(internalMonitor, namespace, serviceAccountName, interalLabels)

		querySFS, destroySFS, ensureInit, checkDBReady, err := statefulset.AdaptFunc(
			internalMonitor,
			namespace,
			sfsName,
			image,
			interalLabels,
			serviceAccountName,
			desiredKind.Spec.ReplicaCount,
			desiredKind.Spec.StorageCapacity,
			cockroachPort,
			cockroachHTTPPort,
			desiredKind.Spec.StorageClass,
			desiredKind.Spec.NodeSelector,
			desiredKind.Spec.Tolerations,
		)
		if err != nil {
			return nil, nil, err
		}

		queryS, destroyS, err := services.AdaptFunc(internalMonitor, namespace, publicServiceName, sfsName, interalLabels, cockroachPort, cockroachHTTPPort)

		//externalName := "cockroachdb-public." + namespaceStr + ".svc.cluster.local"
		//queryES, destroyES, err := service.AdaptFunc("cockroachdb-public", "default", labels, []service.Port{}, "ExternalName", map[string]string{}, false, "", externalName)
		//if err != nil {
		//	return nil, nil, err
		//}

		queryPDB, err := pdb.AdaptFuncToEnsure(namespace, pdbName, interalLabels, "1")
		if err != nil {
			return nil, nil, err
		}

		destroyPDB, err := pdb.AdaptFuncToDestroy(namespace, pdbName)
		if err != nil {
			return nil, nil, err
		}

		currentDB := &Current{
			Common: &tree.Common{
				Kind:    "zitadel.caos.ch/ManagedDatabase",
				Version: "v0",
			},
		}
		current.Parsed = currentDB

		queriers := make([]zitadel.QueryFunc, 0)
		for _, feature := range features {
			if feature == "database" {
				queriers = append(queriers,
					queryRBAC,
					queryCert,
					zitadel.ResourceQueryToZitadelQuery(querySFS),
					zitadel.ResourceQueryToZitadelQuery(queryPDB),
					queryS,
					zitadel.EnsureFuncToQueryFunc(ensureInit),
				)
			}
		}

		destroyers := make([]zitadel.DestroyFunc, 0)
		for _, feature := range features {
			if feature == "database" {
				destroyers = append(destroyers,
					zitadel.ResourceDestroyToZitadelDestroy(destroyPDB),
					destroyS,
					zitadel.ResourceDestroyToZitadelDestroy(destroySFS),
					destroyRBAC,
					destroyCert,
				)
			}
		}

		if desiredKind.Spec.Backups != nil {
			databases := []string{
				"adminapi",
				"auth",
				"authz",
				"eventstore",
				"management",
				"notification",
			}

			oneBackup := false
			for backupName := range desiredKind.Spec.Backups {
				if timestamp != "" && strings.HasPrefix(timestamp, backupName) {
					oneBackup = true
				}
			}

			for backupName, desiredBackup := range desiredKind.Spec.Backups {
				currentBackup := &tree.Tree{}
				if timestamp == "" || !oneBackup || (timestamp != "" && strings.HasPrefix(timestamp, backupName)) {
					queryB, destroyB, err := backups.GetQueryAndDestroyFuncs(
						internalMonitor,
						desiredBackup,
						currentBackup,
						backupName,
						namespace,
						interalLabels,
						databases,
						checkDBReady,
						strings.TrimPrefix(timestamp, backupName+"."),
						secretPasswordName,
						migrationUser,
						users,
						features,
					)
					if err != nil {
						return nil, nil, err
					}

					destroyers = append(destroyers, destroyB)
					queriers = append(queriers, queryB)
				}
			}
		}

		return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
				currentDB.Current.Port = strconv.Itoa(int(cockroachPort))
				currentDB.Current.URL = publicServiceName
				currentDB.Current.ReadyFunc = checkDBReady

				core.SetQueriedForDatabase(queried, current)
				internalMonitor.Info("set current state of managed database")

				ensure, err := zitadel.QueriersToEnsureFunc(internalMonitor, true, queriers, k8sClient, queried)

				return ensure, err
			},
			zitadel.DestroyersToDestroyFunc(internalMonitor, destroyers),
			nil
	}
}
