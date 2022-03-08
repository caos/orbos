package cronjob

import (
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
	"k8s.io/api/batch/v1beta1"
)

func AdaptFuncToEnsure(job *v1beta1.CronJob) (resources.QueryFunc, error) {
	return func(_ kubernetes.ClientInt, _ map[string]interface{}) (resources.EnsureFunc, error) {
		return func(k8sClient kubernetes.ClientInt) error {
			return k8sClient.ApplyCronJob(job)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(namespace, name string) (resources.DestroyFunc, error) {
	return func(client kubernetes.ClientInt) error {
		return client.DeleteCronJob(namespace, name)
	}, nil
}
