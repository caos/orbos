package rbac

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/clusterrole"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/clusterrolebinding"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/role"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/rolebinding"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/serviceaccount"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/mntr"
)

func AdaptFunc(
	monitor mntr.Monitor,
	namespace string,
	name string,
	labels map[string]string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {

	internalMonitor := monitor.WithField("component", "rbac")

	serviceAccountName := name
	roleName := name
	clusterRoleName := name

	destroySA, err := serviceaccount.AdaptFuncToDestroy(namespace, serviceAccountName)
	if err != nil {
		return nil, nil, err
	}

	destroyR, err := role.AdaptFuncToDestroy(roleName, namespace)
	if err != nil {
		return nil, nil, err
	}

	destroyCR, err := clusterrole.AdaptFuncToDestroy(clusterRoleName)
	if err != nil {
		return nil, nil, err
	}

	destroyRB, err := rolebinding.AdaptFuncToDestroy(roleName, namespace)
	if err != nil {
		return nil, nil, err
	}

	destroyCRB, err := clusterrolebinding.AdaptFuncToDestroy(roleName)
	if err != nil {
		return nil, nil, err
	}

	destroyers := []zitadel.DestroyFunc{
		zitadel.ResourceDestroyToZitadelDestroy(destroyR),
		zitadel.ResourceDestroyToZitadelDestroy(destroyCR),
		zitadel.ResourceDestroyToZitadelDestroy(destroyRB),
		zitadel.ResourceDestroyToZitadelDestroy(destroyCRB),
		zitadel.ResourceDestroyToZitadelDestroy(destroySA),
	}

	querySA, err := serviceaccount.AdaptFuncToEnsure(namespace, serviceAccountName, labels)
	if err != nil {
		return nil, nil, err
	}

	queryR, err := role.AdaptFuncToEnsure(roleName, namespace, labels, []string{""}, []string{"secrets"}, []string{"create", "get"})
	if err != nil {
		return nil, nil, err
	}

	queryCR, err := clusterrole.AdaptFuncToEnsure(clusterRoleName, labels, []string{"certificates.k8s.io"}, []string{"certificatesigningrequests"}, []string{"create", "get", "watch"})
	if err != nil {
		return nil, nil, err
	}

	subjects := []rolebinding.Subject{{Kind: "ServiceAccount", Name: serviceAccountName, Namespace: namespace}}
	queryRB, err := rolebinding.AdaptFuncToEnsure(roleName, namespace, labels, subjects, roleName)
	if err != nil {
		return nil, nil, err
	}

	subjectsCRB := []clusterrolebinding.Subject{{Kind: "ServiceAccount", Name: serviceAccountName, Namespace: namespace}}
	queryCRB, err := clusterrolebinding.AdaptFuncToEnsure(roleName, labels, subjectsCRB, roleName)
	if err != nil {
		return nil, nil, err
	}

	queriers := []zitadel.QueryFunc{
		//serviceaccount
		zitadel.ResourceQueryToZitadelQuery(querySA),
		//rbac
		zitadel.ResourceQueryToZitadelQuery(queryR),
		zitadel.ResourceQueryToZitadelQuery(queryCR),
		zitadel.ResourceQueryToZitadelQuery(queryRB),
		zitadel.ResourceQueryToZitadelQuery(queryCRB),
	}
	return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
			return zitadel.QueriersToEnsureFunc(internalMonitor, false, queriers, k8sClient, queried)
		},
		zitadel.DestroyersToDestroyFunc(internalMonitor, destroyers),
		nil

}
