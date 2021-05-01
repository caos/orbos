package secret

import (
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
	"github.com/caos/orbos/pkg/labels"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AdaptFuncToEnsure(namespace string, id labels.IDLabels, data map[string]string) (resources.QueryFunc, error) {

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      id.Name(),
			Namespace: namespace,
			Labels:    labels.MustK8sMap(id),
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: data,
	}
	return func(_ kubernetes.ClientInt) (resources.EnsureFunc, error) {
		return func(k8sClient kubernetes.ClientInt) error {
			return k8sClient.ApplySecret(secret)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(namespace, name string) (resources.DestroyFunc, error) {
	return func(k8sClient kubernetes.ClientInt) error {
		return k8sClient.DeleteSecret(namespace, name)
	}, nil
}
