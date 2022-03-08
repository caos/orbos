package pdb

import (
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
	"github.com/caos/orbos/pkg/labels"
	policy "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func AdaptFuncToEnsure(namespace string, nameLabels *labels.Name, target *labels.Selector, maxUnavailable string) (resources.QueryFunc, error) {
	maxUnavailableParsed := intstr.Parse(maxUnavailable)
	pdb := &policy.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nameLabels.Name(),
			Namespace: namespace,
			Labels:    labels.MustK8sMap(nameLabels),
		},
		Spec: policy.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels.MustK8sMap(target),
			},
			MaxUnavailable: &maxUnavailableParsed,
		},
	}
	return func(_ kubernetes.ClientInt, _ map[string]interface{}) (resources.EnsureFunc, error) {
		return func(k8sClient kubernetes.ClientInt) error {
			return k8sClient.ApplyPodDisruptionBudget(pdb)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(namespace, name string) (resources.DestroyFunc, error) {
	return func(client kubernetes.ClientInt) error {
		return client.DeletePodDisruptionBudget(namespace, name)
	}, nil
}
