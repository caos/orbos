package pdb

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	policy "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func AdaptFunc(namespace, name string, labels map[string]string, maxUnavailable string) (resources.QueryFunc, resources.DestroyFunc, error) {
	return func() (resources.EnsureFunc, error) {
			return func(k8sClient *kubernetes.Client) error {
				maxUnavailableParsed := intstr.Parse(maxUnavailable)

				return k8sClient.ApplyPodDisruptionBudget(&policy.PodDisruptionBudget{
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
				})
			}, nil
		}, func(k8sClient *kubernetes.Client) error {
			//TODO
			return nil
		}, nil
}
