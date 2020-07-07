package managed

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
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

		queriers := make([]resources.QueryFunc, 0)
		destroyers := make([]resources.DestroyFunc, 0)

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

		userList := []string{"root", "flyway", "management", "auth", "authz", "admin", "notify"}
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

		queriers = append(queriers, querySA, queryR, queryCR, queryRB, queryCRB, querySPD, querySP, queryS, querySFS, queryPDB, queryJ)
		destroyCertZ := func(k8sClient *kubernetes.Client) error {
			return destroyCert(k8sClient)
		}
		destroyers = append(destroyers, destroyNS, destroyCertZ, destroySA, destroyR, destroyCR, destroyRB, destroyCRB, destroySPD, destroySP, destroyS, destroySFS, destroyPDB, destroyJ)

		currentDB := &Current{
			Common: &tree.Common{
				Kind:    "zitadel.caos.ch/ManagedDatabase",
				Version: "v0",
			},
		}
		current.Parsed = currentDB

		return func(k8sClient *kubernetes.Client) (zitadel.EnsureFunc, error) {
				currentDB.Current.Port = cockroachPort
				currentDB.Current.URL = cockroachURL

				ensurers := make([]resources.EnsureFunc, 0)
				ensureNS, err := queryNS()
				if err != nil {
					return nil, err
				}
				ensurers = append(ensurers, ensureNS)

				ensureCert, err := queryCert(k8sClient)
				if err != nil {
					return nil, err
				}
				ensurers = append(ensurers, func(k8sClient *kubernetes.Client) error {
					return ensureCert(k8sClient)
				})

				for _, querier := range queriers {
					ensurer, err := querier()
					if err != nil {
						return nil, err
					}
					ensurers = append(ensurers, ensurer)
				}

				return func(k8sClient *kubernetes.Client) error {
					for _, ensurer := range ensurers {
						if err := ensurer(k8sClient); err != nil {
							return err
						}
					}
					return nil
				}, nil
			}, func(k8sClient *kubernetes.Client) error {
				for _, destroyer := range destroyers {
					if err := destroyer(k8sClient); err != nil {
						return err
					}
				}
				return nil
			},
			nil
	}
}
