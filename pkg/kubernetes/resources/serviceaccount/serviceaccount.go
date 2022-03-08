package serviceaccount

import (
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
	"github.com/caos/orbos/pkg/labels"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AdaptFuncToEnsure(namespace string, nameLabels *labels.Name) (resources.QueryFunc, error) {
	sa := &corev1.ServiceAccount{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nameLabels.Name(),
			Namespace: namespace,
			Labels:    labels.MustK8sMap(nameLabels),
		},
	}
	return func(_ kubernetes.ClientInt, _ map[string]interface{}) (resources.EnsureFunc, error) {
		return func(k8sClient kubernetes.ClientInt) error {
			return k8sClient.ApplyServiceAccount(sa)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(namespace, name string) (resources.DestroyFunc, error) {
	return func(client kubernetes.ClientInt) error {
		return client.DeleteServiceAccount(namespace, name)
	}, nil
}
