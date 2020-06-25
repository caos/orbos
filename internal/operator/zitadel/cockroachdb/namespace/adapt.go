package namespace

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/zitadel/cockroachdb"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AdaptFunc(k8sClient *kubernetes.Client, namespace string) (cockroachdb.QueryFunc, cockroachdb.DestroyFunc, error) {
	namespaceList, err := k8sClient.ListNamespaces()
	if err != nil {
		return nil, nil, err
	}

	return func() (cockroachdb.EnsureFunc, error) {
			return func() error {
				found := false
				for _, namespaceItem := range namespaceList.Items {
					if namespaceItem.Name == namespace {
						found = true
					}
				}
				if !found {
					if err := k8sClient.ApplyNamespace(&corev1.Namespace{
						TypeMeta: metav1.TypeMeta{
							Kind:       "Namespace",
							APIVersion: "v1",
						}, ObjectMeta: metav1.ObjectMeta{
							Name: namespace,
						},
					}); err != nil {
						return err
					}
				}
				return nil
			}, nil
		}, func() error {
			//TODO
			return nil
		}, nil
}
