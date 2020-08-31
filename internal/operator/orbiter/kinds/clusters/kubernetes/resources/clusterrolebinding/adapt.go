package clusterrolebinding

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Subject struct {
	Kind      string
	Name      string
	Namespace string
}

func AdaptFuncToEnsure(name string, labels map[string]string, subjects []Subject, clusterrole string) (resources.QueryFunc, error) {
	subjectsList := make([]rbac.Subject, 0)
	for _, subject := range subjects {
		subjectsList = append(subjectsList, rbac.Subject{
			Name:      subject.Name,
			Namespace: subject.Namespace,
			Kind:      subject.Kind,
		})
	}

	crb := &rbac.ClusterRoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:   name,
			Labels: labels,
		},
		Subjects: subjectsList,
		RoleRef: rbac.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Name:     clusterrole,
			Kind:     "ClusterRole",
		},
	}
	return func(_ *kubernetes.Client) (resources.EnsureFunc, error) {
		return func(k8sClient *kubernetes.Client) error {
			return k8sClient.ApplyClusterRoleBinding(crb)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(name string) (resources.DestroyFunc, error) {
	return func(client *kubernetes.Client) error {
		return client.DeleteClusterRoleBinding(name)
	}, nil
}
