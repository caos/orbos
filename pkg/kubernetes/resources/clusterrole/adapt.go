package clusterrole

import (
	kubernetes2 "github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AdaptFuncToEnsure(name string, labels map[string]string, apiGroups, kubeResources, verbs []string) (resources.QueryFunc, error) {
	cr := &rbac.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Rules: []rbac.PolicyRule{{
			APIGroups: apiGroups,
			Resources: kubeResources,
			Verbs:     verbs,
		}},
	}
	return func(_ *kubernetes2.Client) (resources.EnsureFunc, error) {
		return func(k8sClient *kubernetes2.Client) error {
			return k8sClient.ApplyClusterRole(cr)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(name string) (resources.DestroyFunc, error) {
	return func(client *kubernetes2.Client) error {
		return client.DeleteClusterRole(name)
	}, nil
}
