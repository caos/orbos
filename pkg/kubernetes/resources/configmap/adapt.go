package configmap

import (
	kubernetes2 "github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AdaptFuncToEnsure(namespace string, name string, labels map[string]string, data map[string]string) (resources.QueryFunc, error) {
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Data: data,
	}
	return func(_ *kubernetes2.Client) (resources.EnsureFunc, error) {
		return func(k8sClient *kubernetes2.Client) error {
			return k8sClient.ApplyConfigmap(cm)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(namespace, name string) (resources.DestroyFunc, error) {
	return func(client *kubernetes2.Client) error {
		return client.DeleteConfigmap(namespace, name)
	}, nil
}
