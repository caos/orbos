package deployment

import (
	kubernetes2 "github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
	appsv1 "k8s.io/api/apps/v1"
)

func AdaptFuncToEnsure(deployment *appsv1.Deployment) (resources.QueryFunc, error) {
	return func(_ *kubernetes2.Client) (resources.EnsureFunc, error) {
		return func(k8sClient *kubernetes2.Client) error {
			return k8sClient.ApplyDeployment(deployment)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(namespace, name string) (resources.DestroyFunc, error) {
	return func(client *kubernetes2.Client) error {
		return client.DeleteDeployment(namespace, name)
	}, nil
}
