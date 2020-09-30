package role

import (
	kubernetes2 "github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AdaptFuncToEnsure(namespace string, name string, labels map[string]string, apiGroups, kubeResources, verbs []string) (resources.QueryFunc, error) {
	role := &rbac.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Rules: []rbac.PolicyRule{{
			APIGroups: apiGroups,
			Resources: kubeResources,
			Verbs:     verbs,
		}},
	}
	return func(_ *kubernetes2.Client) (resources.EnsureFunc, error) {
		return func(k8sClient *kubernetes2.Client) error {
			return k8sClient.ApplyRole(role)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(name, namespace string) (resources.DestroyFunc, error) {
	return func(client *kubernetes2.Client) error {
		return client.DeleteRole(namespace, name)
	}, nil
}
