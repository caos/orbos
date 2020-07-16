package deployment

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	appsv1 "k8s.io/api/apps/v1"
)

func AdaptFuncToEnsure(deployment *appsv1.Deployment) (resources.QueryFunc, error) {
	return func(_ *kubernetes.Client) (resources.EnsureFunc, error) {
		return func(k8sClient *kubernetes.Client) error {
			if err := k8sClient.ApplyDeployment(deployment); err != nil {
				return err
			}

			return k8sClient.WaitUntilDeploymentReady(deployment.Namespace, deployment.Name)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(name, namespace string) (resources.DestroyFunc, error) {
	return func(client *kubernetes.Client) error {
		//TODO
		return nil
	}, nil
}
