package clusterrole

import (
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
	"github.com/caos/orbos/pkg/labels"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func AdaptFuncToEnsure(nameLabels *labels.Name, apiGroups, kubeResources, verbs []string) (resources.QueryFunc, error) {
	cr := &rbac.ClusterRole{
		ObjectMeta: metav1.ObjectMeta{
			Name:   nameLabels.Name(),
			Labels: labels.MustK8sMap(nameLabels),
		},
		Rules: []rbac.PolicyRule{{
			APIGroups: apiGroups,
			Resources: kubeResources,
			Verbs:     verbs,
		}},
	}
	return func(_ kubernetes.ClientInt) (resources.EnsureFunc, error) {
		return func(k8sClient kubernetes.ClientInt) error {
			return k8sClient.ApplyClusterRole(cr)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(name string) (resources.DestroyFunc, error) {
	return func(client kubernetes.ClientInt) error {
		return client.DeleteClusterRole(name)
	}, nil
}
