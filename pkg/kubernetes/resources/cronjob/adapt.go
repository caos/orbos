package cronjob

import (
	kubernetes2 "github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
	"k8s.io/api/batch/v1beta1"
)

func AdaptFuncToEnsure(job *v1beta1.CronJob) (resources.QueryFunc, error) {
	return func(_ *kubernetes2.Client) (resources.EnsureFunc, error) {
		return func(k8sClient *kubernetes2.Client) error {
			return k8sClient.ApplyCronJob(job)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(namespace, name string) (resources.DestroyFunc, error) {
	return func(client *kubernetes2.Client) error {
		return client.DeleteCronJob(namespace, name)
	}, nil
}
