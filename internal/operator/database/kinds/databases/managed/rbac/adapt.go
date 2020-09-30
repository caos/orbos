package rbac

import (
	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/mntr"
	kubernetes2 "github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources/clusterrole"
	"github.com/caos/orbos/pkg/kubernetes/resources/clusterrolebinding"
	"github.com/caos/orbos/pkg/kubernetes/resources/role"
	"github.com/caos/orbos/pkg/kubernetes/resources/rolebinding"
	"github.com/caos/orbos/pkg/kubernetes/resources/serviceaccount"
)

func AdaptFunc(
	monitor mntr.Monitor,
	namespace string,
	name string,
	labels map[string]string,
) (
	core.QueryFunc,
	core.DestroyFunc,
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

	destroyR, err := role.AdaptFuncToDestroy(namespace, roleName)
	if err != nil {
		return nil, nil, err
	}

	destroyCR, err := clusterrole.AdaptFuncToDestroy(clusterRoleName)
	if err != nil {
		return nil, nil, err
	}

	destroyRB, err := rolebinding.AdaptFuncToDestroy(namespace, roleName)
	if err != nil {
		return nil, nil, err
	}

	destroyCRB, err := clusterrolebinding.AdaptFuncToDestroy(roleName)
	if err != nil {
		return nil, nil, err
	}

	destroyers := []core.DestroyFunc{
		core.ResourceDestroyToZitadelDestroy(destroyR),
		core.ResourceDestroyToZitadelDestroy(destroyCR),
		core.ResourceDestroyToZitadelDestroy(destroyRB),
		core.ResourceDestroyToZitadelDestroy(destroyCRB),
		core.ResourceDestroyToZitadelDestroy(destroySA),
	}

	querySA, err := serviceaccount.AdaptFuncToEnsure(namespace, serviceAccountName, labels)
	if err != nil {
		return nil, nil, err
	}

	queryR, err := role.AdaptFuncToEnsure(namespace, roleName, labels, []string{""}, []string{"secrets"}, []string{"create", "get"})
	if err != nil {
		return nil, nil, err
	}

	queryCR, err := clusterrole.AdaptFuncToEnsure(clusterRoleName, labels, []string{"certificates.k8s.io"}, []string{"certificatesigningrequests"}, []string{"create", "get", "watch"})
	if err != nil {
		return nil, nil, err
	}

	subjects := []rolebinding.Subject{{Kind: "ServiceAccount", Name: serviceAccountName, Namespace: namespace}}
	queryRB, err := rolebinding.AdaptFuncToEnsure(namespace, roleName, labels, subjects, roleName)
	if err != nil {
		return nil, nil, err
	}

	subjectsCRB := []clusterrolebinding.Subject{{Kind: "ServiceAccount", Name: serviceAccountName, Namespace: namespace}}
	queryCRB, err := clusterrolebinding.AdaptFuncToEnsure(roleName, labels, subjectsCRB, roleName)
	if err != nil {
		return nil, nil, err
	}

	queriers := []core.QueryFunc{
		//serviceaccount
		core.ResourceQueryToZitadelQuery(querySA),
		//rbac
		core.ResourceQueryToZitadelQuery(queryR),
		core.ResourceQueryToZitadelQuery(queryCR),
		core.ResourceQueryToZitadelQuery(queryRB),
		core.ResourceQueryToZitadelQuery(queryCRB),
	}
	return func(k8sClient *kubernetes2.Client, queried map[string]interface{}) (core.EnsureFunc, error) {
			return core.QueriersToEnsureFunc(internalMonitor, false, queriers, k8sClient, queried)
		},
		core.DestroyersToDestroyFunc(internalMonitor, destroyers),
		nil

}
