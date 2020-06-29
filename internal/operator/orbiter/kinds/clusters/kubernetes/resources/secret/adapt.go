package secret

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AdaptFunc(k8sClient *kubernetes.Client, name string, namespace string, labels map[string]string, data map[string]string) (resources.QueryFunc, resources.DestroyFunc, error) {
	return func() (resources.EnsureFunc, error) {
			return func() error {
				return k8sClient.ApplySecret(&corev1.Secret{
					ObjectMeta: metav1.ObjectMeta{
						Name:      name,
						Namespace: namespace,
						Labels:    labels,
					},
					Type:       corev1.SecretTypeOpaque,
					StringData: data,
				})
			}, nil
		}, func() error {
			//TODO
			return nil
		}, nil
}
