package statefulset

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	appsv1 "k8s.io/api/apps/v1"
)

func AdaptFunc(k8sClient *kubernetes.Client, statefulset *appsv1.StatefulSet) (resources.QueryFunc, resources.DestroyFunc, error) {
	return func() (resources.EnsureFunc, error) {
			return func() error {
				return k8sClient.ApplyStatefulSet(statefulset)
			}, nil
		}, func() error {
			//TODO
			return nil
		}, nil
}
