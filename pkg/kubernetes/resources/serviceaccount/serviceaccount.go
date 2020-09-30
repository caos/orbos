package serviceaccount

import (
	kubernetes2 "github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
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
	return func(_ *kubernetes2.Client) (resources.EnsureFunc, error) {
		return func(k8sClient *kubernetes2.Client) error {
			return k8sClient.ApplyServiceAccount(sa)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(namespace, name string) (resources.DestroyFunc, error) {
	return func(client *kubernetes2.Client) error {
		return client.DeleteServiceAccount(namespace, name)
	}, nil
}
