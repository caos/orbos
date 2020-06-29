package job

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	batch "k8s.io/api/batch/v1"
)

func AdaptFunc(k8sClient *kubernetes.Client, job *batch.Job) (resources.QueryFunc, resources.DestroyFunc, error) {
	return func() (resources.EnsureFunc, error) {
			return func() error {
				return k8sClient.ApplyJob(job)
			}, nil
		}, func() error {
			//TODO
			return nil
		}, nil
}
