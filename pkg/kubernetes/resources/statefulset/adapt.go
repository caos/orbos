package statefulset

import (
	"github.com/caos/orbos/v5/pkg/kubernetes"
	"github.com/caos/orbos/v5/pkg/kubernetes/resources"
	appsv1 "k8s.io/api/apps/v1"
)

func AdaptFuncToEnsure(statefulset *appsv1.StatefulSet, force bool) (resources.QueryFunc, error) {
	return func(_ kubernetes.ClientInt) (resources.EnsureFunc, error) {
		return func(k8sClient kubernetes.ClientInt) error {
			return k8sClient.ApplyStatefulSet(statefulset, force)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(namespace, name string) (resources.DestroyFunc, error) {
	return func(k8sClient kubernetes.ClientInt) error {
		return k8sClient.DeleteStatefulset(namespace, name)
	}, nil
}
