package namespace

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AdaptFunc(k8sClient *kubernetes.Client, namespace string) (resources.QueryFunc, resources.DestroyFunc, error) {
	return func() (resources.EnsureFunc, error) {
			return func() error {
				return k8sClient.ApplyNamespace(&corev1.Namespace{
					ObjectMeta: metav1.ObjectMeta{
						Name: namespace,
					},
				})
			}, nil
		}, func() error {
			//TODO
			return nil
		}, nil
}
