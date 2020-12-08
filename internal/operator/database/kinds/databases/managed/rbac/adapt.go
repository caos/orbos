package rbac

import (
	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources/clusterrole"
	"github.com/caos/orbos/pkg/kubernetes/resources/clusterrolebinding"
	"github.com/caos/orbos/pkg/kubernetes/resources/role"
	"github.com/caos/orbos/pkg/kubernetes/resources/rolebinding"
	"github.com/caos/orbos/pkg/kubernetes/resources/serviceaccount"
	"github.com/caos/orbos/pkg/labels"
)

func AdaptFunc(
	monitor mntr.Monitor,
	namespace string,
	nameLabels *labels.Name,
) (
	core.QueryFunc,
	core.DestroyFunc,
	error,
) {

	internalMonitor := monitor.WithField("component", "rbac")

	serviceAccountLabels := nameLabels
	roleLabels := nameLabels
	clusterRoleLabels := nameLabels

	destroySA, err := serviceaccount.AdaptFuncToDestroy(namespace, serviceAccountLabels.Name())
	if err != nil {
		return nil, nil, err
	}

	destroyR, err := role.AdaptFuncToDestroy(namespace, roleLabels.Name())
	if err != nil {
		return nil, nil, err
	}

	destroyCR, err := clusterrole.AdaptFuncToDestroy(clusterRoleLabels.Name())
	if err != nil {
		return nil, nil, err
	}

	destroyRB, err := rolebinding.AdaptFuncToDestroy(namespace, roleLabels.Name())
	if err != nil {
		return nil, nil, err
	}

	destroyCRB, err := clusterrolebinding.AdaptFuncToDestroy(roleLabels.Name())
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

	querySA, err := serviceaccount.AdaptFuncToEnsure(namespace, serviceAccountLabels)
	if err != nil {
		return nil, nil, err
	}

	queryR, err := role.AdaptFuncToEnsure(namespace, roleLabels, []string{""}, []string{"secrets"}, []string{"create", "get"})
	if err != nil {
		return nil, nil, err
	}

	queryCR, err := clusterrole.AdaptFuncToEnsure(clusterRoleLabels, []string{"certificates.k8s.io"}, []string{"certificatesigningrequests"}, []string{"create", "get", "watch"})
	if err != nil {
		return nil, nil, err
	}

	subjects := []rolebinding.Subject{{Kind: "ServiceAccount", Name: serviceAccountLabels.Name(), Namespace: namespace}}
	queryRB, err := rolebinding.AdaptFuncToEnsure(namespace, roleLabels, subjects, roleLabels.Name())
	if err != nil {
		return nil, nil, err
	}

	subjectsCRB := []clusterrolebinding.Subject{{Kind: "ServiceAccount", Name: serviceAccountLabels.Name(), Namespace: namespace}}
	queryCRB, err := clusterrolebinding.AdaptFuncToEnsure(roleLabels, subjectsCRB, roleLabels.Name())
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
	return func(k8sClient kubernetes.ClientInt, queried map[string]interface{}) (core.EnsureFunc, error) {
			return core.QueriersToEnsureFunc(internalMonitor, false, queriers, k8sClient, queried)
		},
		core.DestroyersToDestroyFunc(internalMonitor, destroyers),
		nil

}
