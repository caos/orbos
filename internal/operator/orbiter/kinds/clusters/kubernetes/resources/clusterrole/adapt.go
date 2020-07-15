package clusterrole

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AdaptFunc(name string, labels map[string]string, apiGroups, kubeResources, verbs []string) (resources.QueryFunc, resources.DestroyFunc, error) {
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
	return func(_ *kubernetes.Client) (resources.EnsureFunc, error) {
			return func(k8sClient *kubernetes.Client) error {
				return k8sClient.ApplyClusterRole(cr)
			}, nil
		}, func(k8sClient *kubernetes.Client) error {
			//TODO
			return nil
		}, nil
}
