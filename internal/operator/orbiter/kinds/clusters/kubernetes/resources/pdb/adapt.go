package pdb

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
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
	return func(_ *kubernetes.Client) (resources.EnsureFunc, error) {
		return func(k8sClient *kubernetes.Client) error {
			return k8sClient.ApplyPodDisruptionBudget(pdb)
		}, nil
	}, nil
}

func AdaptFuncToDestroy(name, namespace string) (resources.DestroyFunc, error) {
	return func(client *kubernetes.Client) error {
		//TODO
		return nil
	}, nil
}
