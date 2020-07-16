package rolebinding

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

func AdaptFuncToEnsure(name string, namespace string, labels map[string]string, subjects []Subject, role string) (resources.QueryFunc, error) {
	subjectsList := make([]rbac.Subject, 0)
	for _, subject := range subjects {
		subjectsList = append(subjectsList, rbac.Subject{
			Kind:      subject.Kind,
			Name:      subject.Name,
			Namespace: subject.Namespace,
		})
	}

	rolebinding := &rbac.RoleBinding{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
			Labels:    labels,
		},
		Subjects: subjectsList,
		RoleRef: rbac.RoleRef{
			Name:     role,
			Kind:     "Role",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	return func(_ *kubernetes.Client) (resources.EnsureFunc, error) {
		return func(k8sClient *kubernetes.Client) error {
			return k8sClient.ApplyRoleBinding(rolebinding)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(name, namespace string) (resources.DestroyFunc, error) {
	return func(client *kubernetes.Client) error {
		//TODO
		return nil
	}, nil
}
