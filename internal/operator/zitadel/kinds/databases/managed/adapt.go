package managed

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/pdb"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/managed/certificate"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/managed/initjob"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/managed/rbac"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/managed/services"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/managed/statefulset"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	"strconv"
)

func AdaptFunc(labels map[string]string, users []string, namespace string) zitadel.AdaptFunc {
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
		initJobName := sfsName + "-init"
		serviceAccountName := sfsName
		publicServiceName := sfsName + "-public"
		cockroachPort := int32(26257)
		cockroachHTTPPort := int32(8080)
		image := "cockroachdb/cockroach:v20.1.2"

		userList := []string{"root"}
		userList = append(userList, users...)

		queryCert, destroyCert, err := certificate.AdaptFunc(internalMonitor, namespace, userList, interalLabels, desiredKind.Spec.ClusterDns)
		if err != nil {
			return nil, nil, err
		}

		queryRBAC, destroyRBAC, err := rbac.AdaptFunc(internalMonitor, namespace, serviceAccountName, interalLabels)

		querySFS, destroySFS, checkDBRunning, checkDBReady, err := statefulset.AdaptFunc(internalMonitor, namespace, sfsName, image, interalLabels, serviceAccountName, desiredKind.Spec.ReplicaCount, desiredKind.Spec.StorageCapacity, cockroachPort, cockroachHTTPPort, desiredKind.Spec.StorageClass, desiredKind.Spec.NodeSelector)
		if err != nil {
			return nil, nil, err
		}

		queryS, destroyS, err := services.AdaptFunc(internalMonitor, namespace, publicServiceName, sfsName, interalLabels, cockroachPort, cockroachHTTPPort)

		//externalName := "cockroachdb-public." + namespaceStr + ".svc.cluster.local"
		//queryES, destroyES, err := service.AdaptFunc("cockroachdb-public", "default", labels, []service.Port{}, "ExternalName", map[string]string{}, false, "", externalName)
		//if err != nil {
		//	return nil, nil, err
		//}

		queryJ, destroyJ, err := initjob.AdaptFunc(internalMonitor, namespace, initJobName, image, labels, serviceAccountName, checkDBRunning)
		if err != nil {
			return nil, nil, err
		}

		queryPDB, err := pdb.AdaptFuncToEnsure(namespace, pdbName, interalLabels, "1")
		if err != nil {
			return nil, nil, err
		}

		queriers := []zitadel.QueryFunc{
			queryRBAC,
			queryCert,
			zitadel.ResourceQueryToZitadelQuery(querySFS),
			zitadel.ResourceQueryToZitadelQuery(queryPDB),
			queryS,
			queryJ,
		}

		destroyPDB, err := pdb.AdaptFuncToDestroy(namespace, pdbName)
		if err != nil {
			return nil, nil, err
		}

		destroyers := []zitadel.DestroyFunc{
			destroyJ,
			zitadel.ResourceDestroyToZitadelDestroy(destroyPDB),
			destroyS,
			zitadel.ResourceDestroyToZitadelDestroy(destroySFS),
			destroyRBAC,
			destroyCert,
		}

		currentDB := &Current{
			Common: &tree.Common{
				Kind:    "zitadel.caos.ch/ManagedDatabase",
				Version: "v0",
			},
		}
		current.Parsed = currentDB

		return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
				ensure, err := zitadel.QueriersToEnsureFunc(internalMonitor, true, queriers, k8sClient, queried)

				currentDB.Current.Port = strconv.Itoa(int(cockroachPort))
				currentDB.Current.URL = publicServiceName
				currentDB.Current.ReadyFunc = checkDBReady

				internalMonitor.Info("set current state of managed database")

				return ensure, err
			},
			zitadel.DestroyersToDestroyFunc(internalMonitor, destroyers),
			nil
	}
}
