package pdb

import (
	kubernetes2 "github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources"
	policy "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func AdaptFuncToEnsure(namespace, name string, labels map[string]string, maxUnavailable string) (resources.QueryFunc, error) {
	maxUnavailableParsed := intstr.Parse(maxUnavailable)
	pdb := &policy.PodDisruptionBudget{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: policy.PodDisruptionBudgetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			MaxUnavailable: &maxUnavailableParsed,
		},
	}
	return func(_ *kubernetes2.Client) (resources.EnsureFunc, error) {
		return func(k8sClient *kubernetes2.Client) error {
			return k8sClient.ApplyPodDisruptionBudget(pdb)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(namespace, name string) (resources.DestroyFunc, error) {
	return func(client *kubernetes2.Client) error {
		return client.DeletePodDisruptionBudget(namespace, name)
	}, nil
}
