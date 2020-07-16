package serviceaccount

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AdaptFuncToEnsure(namespace, name string, labels map[string]string) (resources.QueryFunc, error) {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
	}
	return func(_ *kubernetes.Client) (resources.EnsureFunc, error) {
		return func(k8sClient *kubernetes.Client) error {
			return k8sClient.ApplyServiceAccount(sa)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(name, namespace string) (resources.DestroyFunc, error) {
	return func(client *kubernetes.Client) error {
		//TODO
		return nil
	}, nil
}
