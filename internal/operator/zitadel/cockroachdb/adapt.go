package cockroachdb

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
	"github.com/caos/orbos/internal/operator/zitadel/cockroachdb/certificate"
	"github.com/caos/orbos/internal/operator/zitadel/cockroachdb/initjob"
	"github.com/caos/orbos/internal/operator/zitadel/cockroachdb/statefulset"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	"io/ioutil"
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

		data, err := ioutil.ReadFile("/Users/benz/.kube/config")
		dummyKubeconfig := string(data)
		k8sClient := kubernetes.NewK8sClient(monitor, &dummyKubeconfig)
		//if err := k8sClient.RefreshLocal(); err != nil {
		//	return nil, nil, err
		//}

		if !k8sClient.Available() {
			return nil, nil, errors.New("kubeconfig failed")
		}
		queriers := make([]resources.QueryFunc, 0)
		destroyers := make([]resources.DestroyFunc, 0)

		namespaceStr := "caos-cockroach"
		labels := map[string]string{"app.kubernetes.io/managed-by": "zitadel.caos.ch"}
		serviceAccountName := "cockroachdb"
		roleName := "cockroachdb"
		clusterRoleName := "cockroachdb"

		queryNS, destroyNS, err := namespace.AdaptFunc(k8sClient, namespaceStr)
		if err != nil {
			return nil, nil, err
		}

		userList := []string{"root"}
		if desiredKind.Spec.Users != nil && len(desiredKind.Spec.Users) > 0 {
			userList = append(userList, desiredKind.Spec.Users...)
		}
		queryCert, destroyCert, err := certificate.AdaptFunc(k8sClient, namespaceStr, userList, labels)
		if err != nil {
			return nil, nil, err
		}

		querySA, destroySA, err := serviceaccount.AdaptFunc(k8sClient, namespaceStr, serviceAccountName, labels)
		if err != nil {
			return nil, nil, err
		}

		queryR, destroyR, err := role.AdaptFunc(k8sClient, roleName, namespaceStr, labels, []string{""}, []string{"secrets"}, []string{"create", "get"})
		if err != nil {
			return nil, nil, err
		}

		queryCR, destroyCR, err := clusterrole.AdaptFunc(k8sClient, clusterRoleName, labels, []string{"certificates.k8s.io"}, []string{"certificatesigningrequests"}, []string{"create", "get", "watch"})
		if err != nil {
			return nil, nil, err
		}

		subjects := []rolebinding.Subject{{Kind: "ServiceAccount", Name: serviceAccountName, Namespace: namespaceStr}}
		queryRB, destroyRB, err := rolebinding.AdaptFunc(k8sClient, roleName, namespaceStr, labels, subjects, roleName)
		if err != nil {
			return nil, nil, err
		}

		subjectsCRB := []clusterrolebinding.Subject{{Kind: "ServiceAccount", Name: serviceAccountName, Namespace: namespaceStr}}
		queryCRB, destroyCRB, err := clusterrolebinding.AdaptFunc(k8sClient, roleName, labels, subjectsCRB, roleName)
		if err != nil {
			return nil, nil, err
		}

		ports := []service.Port{
			{Port: 26257, TargetPort: "26257", Name: "grpc"},
			{Port: 8080, TargetPort: "8080", Name: "http"},
		}
		querySP, destroySP, err := service.AdaptFunc(k8sClient, "cockroachdb-public", namespaceStr, labels, ports, "", labels, false, "", "")
		if err != nil {
			return nil, nil, err
		}

		queryS, destroyS, err := service.AdaptFunc(k8sClient, "cockroachdb", namespaceStr, labels, ports, "", labels, true, "None", "")
		if err != nil {
			return nil, nil, err
		}

		querySFS, destroySFS, err := statefulset.AdaptFunc(k8sClient, namespaceStr, labels, serviceAccountName, desiredKind.Spec.ReplicaCount, desiredKind.Spec.StorageCapacity)

		queryPDB, destroyPDB, err := pdb.AdaptFunc(k8sClient, namespaceStr, "cockroachdb-budget", labels, "1")
		if err != nil {
			return nil, nil, err
		}

		externalName := "cockroachdb-public." + namespaceStr + ".svc.cluster.local"
		queryES, destroyES, err := service.AdaptFunc(k8sClient, "cockroachdb-public", "default", labels, []service.Port{}, "ExternalName", map[string]string{}, false, "", externalName)
		if err != nil {
			return nil, nil, err
		}

		queryJ, destroyJ, err := initjob.AdaptFunc(k8sClient, namespaceStr, labels, serviceAccountName)
		if err != nil {
			return nil, nil, err
		}

		queriers = append(queriers, queryNS, queryCert, querySA, queryR, queryCR, queryRB, queryCRB, querySP, queryS, querySFS, queryPDB, queryES, queryJ)
		destroyers = append(destroyers, destroyNS, destroyCert, destroySA, destroyR, destroyCR, destroyRB, destroyCRB, destroySP, destroyS, destroySFS, destroyPDB, destroyES, destroyJ)

		return func() (zitadel.EnsureFunc, error) {
				ensurers := make([]resources.EnsureFunc, 0)
				for _, querier := range queriers {
					ensurer, err := querier()
					if err != nil {
						return nil, err
					}
					ensurers = append(ensurers, ensurer)
				}

				return func() error {
					for _, ensurer := range ensurers {
						if err := ensurer(); err != nil {
							return err
						}
					}
					return nil
				}, nil
			}, func() error {
				for _, destroyer := range destroyers {
					if err := destroyer(); err != nil {
						return err
					}
				}
				return nil
			},
			nil
	}
}
