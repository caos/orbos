package namespace

import (
	kubernetes2 "github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AdaptFuncToEnsure(namespace string) (resources.QueryFunc, error) {
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: namespace,
		},
	}
	return func(_ *kubernetes2.Client) (resources.EnsureFunc, error) {
		return func(k8sClient *kubernetes2.Client) error {
			return k8sClient.ApplyNamespace(ns)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(namespace string) (resources.DestroyFunc, error) {
	return func(client *kubernetes2.Client) error {
		return client.DeleteNamespace(namespace)
	}, nil
}
