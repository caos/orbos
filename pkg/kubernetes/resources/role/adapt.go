package role

import (
	"github.com/caos/orbos/v5/pkg/kubernetes"
	"github.com/caos/orbos/v5/pkg/kubernetes/resources"
	"github.com/caos/orbos/v5/pkg/labels"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AdaptFuncToEnsure(namespace string, nameLabels *labels.Name, apiGroups, kubeResources, verbs []string) (resources.QueryFunc, error) {
	role := &rbac.Role{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nameLabels.Name(),
			Namespace: namespace,
			Labels:    labels.MustK8sMap(nameLabels),
		},
		Rules: []rbac.PolicyRule{{
			APIGroups: apiGroups,
			Resources: kubeResources,
			Verbs:     verbs,
		}},
	}
	return func(_ kubernetes.ClientInt) (resources.EnsureFunc, error) {
		return func(k8sClient kubernetes.ClientInt) error {
			return k8sClient.ApplyRole(role)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(name, namespace string) (resources.DestroyFunc, error) {
	return func(client kubernetes.ClientInt) error {
		return client.DeleteRole(namespace, name)
	}, nil
}
