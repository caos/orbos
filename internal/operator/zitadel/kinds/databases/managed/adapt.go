package managed

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/clusterrole"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/clusterrolebinding"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/namespace"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/pdb"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/role"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/rolebinding"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/service"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/serviceaccount"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/managed/certificate"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/managed/initjob"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/databases/managed/statefulset"
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

		namespaceStr := "caos-zitadel"

		labels := map[string]string{
			"app.kubernetes.io/managed-by": "zitadel.caos.ch",
			"app.kubernetes.io/part-of":    "zitadel",
			"zitadel.caos.ch/part-of":      "database",
		}

		serviceAccountName := "cockroachdb"
		roleName := "cockroachdb"
		clusterRoleName := "cockroachdb"

		queryNS, destroyNS, err := namespace.AdaptFunc(namespaceStr)
		if err != nil {
			return nil, nil, err
		}

		userList := []string{"root", "flyway", "management", "auth", "authz", "adminapi", "notification"}
		queryCert, destroyCert, err := certificate.AdaptFunc(namespaceStr, userList, labels)
		if err != nil {
			return nil, nil, err
		}

		querySA, destroySA, err := serviceaccount.AdaptFunc(namespaceStr, serviceAccountName, labels)
		if err != nil {
			return nil, nil, err
		}

		queryR, destroyR, err := role.AdaptFunc(roleName, namespaceStr, labels, []string{""}, []string{"secrets"}, []string{"create", "get"})
		if err != nil {
			return nil, nil, err
		}

		queryCR, destroyCR, err := clusterrole.AdaptFunc(clusterRoleName, labels, []string{"certificates.k8s.io"}, []string{"certificatesigningrequests"}, []string{"create", "get", "watch"})
		if err != nil {
			return nil, nil, err
		}

		subjects := []rolebinding.Subject{{Kind: "ServiceAccount", Name: serviceAccountName, Namespace: namespaceStr}}
		queryRB, destroyRB, err := rolebinding.AdaptFunc(roleName, namespaceStr, labels, subjects, roleName)
		if err != nil {
			return nil, nil, err
		}

		subjectsCRB := []clusterrolebinding.Subject{{Kind: "ServiceAccount", Name: serviceAccountName, Namespace: namespaceStr}}
		queryCRB, destroyCRB, err := clusterrolebinding.AdaptFunc(roleName, labels, subjectsCRB, roleName)
		if err != nil {
			return nil, nil, err
		}

		cockroachPort := "26257"
		cockroachURL := "cockroachdb-public"
		ports := []service.Port{
			{Port: 26257, TargetPort: cockroachPort, Name: "grpc"},
			{Port: 8080, TargetPort: "8080", Name: "http"},
		}
		querySPD, destroySPD, err := service.AdaptFunc(cockroachURL, "default", labels, ports, "", labels, false, "", "")
		if err != nil {
			return nil, nil, err
		}

		querySP, destroySP, err := service.AdaptFunc(cockroachURL, namespaceStr, labels, ports, "", labels, false, "", "")
		if err != nil {
			return nil, nil, err
		}

		queryS, destroyS, err := service.AdaptFunc("cockroachdb", namespaceStr, labels, ports, "", labels, true, "None", "")
		if err != nil {
			return nil, nil, err
		}

		querySFS, destroySFS, err := statefulset.AdaptFunc(namespaceStr, labels, serviceAccountName, desiredKind.Spec.ReplicaCount, desiredKind.Spec.StorageCapacity)

		queryPDB, destroyPDB, err := pdb.AdaptFunc(namespaceStr, "cockroachdb-budget", labels, "1")
		if err != nil {
			return nil, nil, err
		}

		//externalName := "cockroachdb-public." + namespaceStr + ".svc.cluster.local"
		//queryES, destroyES, err := service.AdaptFunc("cockroachdb-public", "default", labels, []service.Port{}, "ExternalName", map[string]string{}, false, "", externalName)
		//if err != nil {
		//	return nil, nil, err
		//}

		queryJ, destroyJ, err := initjob.AdaptFunc(namespaceStr, labels, serviceAccountName)
		if err != nil {
			return nil, nil, err
		}

		queriers := []zitadel.QueryFunc{
			//namespace
			zitadel.ResourceQueryToZitadelQuery(queryNS),
			//serviceaccount
			zitadel.ResourceQueryToZitadelQuery(querySA),
			//rbac
			zitadel.ResourceQueryToZitadelQuery(queryR),
			zitadel.ResourceQueryToZitadelQuery(queryCR),
			zitadel.ResourceQueryToZitadelQuery(queryRB),
			zitadel.ResourceQueryToZitadelQuery(queryCRB),
			//services
			zitadel.ResourceQueryToZitadelQuery(querySPD),
			zitadel.ResourceQueryToZitadelQuery(querySP),
			zitadel.ResourceQueryToZitadelQuery(queryS),
			//certificates
			queryCert,
			//statefulset
			zitadel.ResourceQueryToZitadelQuery(querySFS),
			//poddisruptionpolicy
			zitadel.ResourceQueryToZitadelQuery(queryPDB),
			//initjob
			zitadel.ResourceQueryToZitadelQuery(queryJ),
		}

		destroyers := []zitadel.DestroyFunc{
			zitadel.ResourceDestroyToZitadelDestroy(destroyJ),
			zitadel.ResourceDestroyToZitadelDestroy(destroyPDB),
			zitadel.ResourceDestroyToZitadelDestroy(destroySPD),
			zitadel.ResourceDestroyToZitadelDestroy(destroySP),
			zitadel.ResourceDestroyToZitadelDestroy(destroyS),
			zitadel.ResourceDestroyToZitadelDestroy(destroySFS),
			zitadel.ResourceDestroyToZitadelDestroy(destroyR),
			zitadel.ResourceDestroyToZitadelDestroy(destroyCR),
			zitadel.ResourceDestroyToZitadelDestroy(destroyRB),
			zitadel.ResourceDestroyToZitadelDestroy(destroyCRB),
			zitadel.ResourceDestroyToZitadelDestroy(destroySA),
			destroyCert,
			zitadel.ResourceDestroyToZitadelDestroy(destroyNS),
		}

		currentDB := &Current{
			Common: &tree.Common{
				Kind:    "zitadel.caos.ch/ManagedDatabase",
				Version: "v0",
			},
		}
		current.Parsed = currentDB

		return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
				currentDB.Current.Port = cockroachPort
				currentDB.Current.URL = cockroachURL

				return zitadel.QueriersToEnsureFunc(queriers, k8sClient, queried)
			},
			zitadel.DestroyersToDestroyFunc(destroyers),
			nil
	}
}
