package job

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	batch "k8s.io/api/batch/v1"
)

func AdaptFunc(job *batch.Job) (resources.QueryFunc, resources.DestroyFunc, error) {
	return func() (resources.EnsureFunc, error) {
			return func(k8sClient *kubernetes.Client) error {
				if err := k8sClient.ApplyJob(job); err != nil {
					return err
				}

				return k8sClient.WaitUntilJobCompleted(job.Namespace, job.Name)
			}, nil
		}, func(k8sClient *kubernetes.Client) error {
			//TODO
			return nil
		}, nil
}
