package rolebinding

import (
	"github.com/caos/orbos/v5/pkg/kubernetes"
	"github.com/caos/orbos/v5/pkg/kubernetes/resources"
	"github.com/caos/orbos/v5/pkg/labels"
	rbac "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Subject struct {
	Kind      string
	Name      string
	Namespace string
}

func AdaptFuncToEnsure(namespace string, nameLabels *labels.Name, subjects []Subject, role string) (resources.QueryFunc, error) {
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
			Name:      nameLabels.Name(),
			Namespace: namespace,
			Labels:    labels.MustK8sMap(nameLabels),
		},
		Subjects: subjectsList,
		RoleRef: rbac.RoleRef{
			Name:     role,
			Kind:     "Role",
			APIGroup: "rbac.authorization.k8s.io",
		},
	}
	return func(_ kubernetes.ClientInt) (resources.EnsureFunc, error) {
		return func(k8sClient kubernetes.ClientInt) error {
			return k8sClient.ApplyRoleBinding(rolebinding)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(namespace, name string) (resources.DestroyFunc, error) {
	return func(client kubernetes.ClientInt) error {
		return client.DeleteRoleBinding(namespace, name)
	}, nil
}
