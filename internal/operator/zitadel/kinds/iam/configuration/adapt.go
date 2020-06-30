package configuration

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
)

func AdaptFunc(
	k8sClient *kubernetes.Client,
	namespace string,
	labels map[string]string,
	migrationConfigmap string,
) (
	resources.QueryFunc,
	resources.DestroyFunc,
	error,
) {

	queriers := make([]resources.QueryFunc, 0)
	destroyers := make([]resources.DestroyFunc, 0)

	//queriers = append(queriers, queryCM)
	//destroyers = append(destroyers, destroyCM)

	return func() (resources.EnsureFunc, error) {
			ensurers := make([]resources.EnsureFunc, 0)
			for _, querier := range queriers {
				ensurer, err := querier()
				if err != nil {
					return nil, err
				}
				ensurers = append(ensurers, ensurer)
			}

			return func(k8sClient *kubernetes.Client) error {
				for _, ensurer := range ensurers {
					if err := ensurer(k8sClient); err != nil {
						return err
					}
				}
				return nil
			}, nil
		}, func(k8sClient *kubernetes.Client) error {
			for _, destroyer := range destroyers {
				if err := destroyer(k8sClient); err != nil {
					return err
				}
			}
			return nil
		},
		nil
}
