package statefulset

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	appsv1 "k8s.io/api/apps/v1"
)

func AdaptFunc(statefulset *appsv1.StatefulSet) (resources.QueryFunc, resources.DestroyFunc, error) {
	return func(_ *kubernetes.Client) (resources.EnsureFunc, error) {
			return func(k8sClient *kubernetes.Client) error {
				if err := k8sClient.ApplyStatefulSet(statefulset); err != nil {
					return err
				}
				return k8sClient.WaitUntilStatefulsetIsReady(statefulset.Namespace, statefulset.Name, true, false)
			}, nil
		}, func(k8sClient *kubernetes.Client) error {
			//TODO
			return nil
		}, nil
}
