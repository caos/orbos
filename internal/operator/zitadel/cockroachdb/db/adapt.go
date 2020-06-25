package db

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/zitadel/cockroachdb"
	"github.com/caos/orbos/internal/operator/zitadel/cockroachdb/certificate"
	"github.com/caos/orbos/internal/operator/zitadel/cockroachdb/namespace"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
)

func AdaptFunc() cockroachdb.AdaptFunc {
	return func(
		monitor mntr.Monitor,
		desired *tree.Tree,
		current *tree.Tree,
	) (
		cockroachdb.QueryFunc,
		cockroachdb.DestroyFunc,
		error,
	) {
		dummyKubeconfig := ""
		k8sClient := kubernetes.NewK8sClient(monitor, &dummyKubeconfig)
		if err := k8sClient.RefreshLocal(); err != nil {
			return nil, nil, err
		}

		queriers := make([]cockroachdb.QueryFunc, 0)
		destroyers := make([]cockroachdb.DestroyFunc, 0)

		queryNS, destroyNS, err := namespace.AdaptFunc(k8sClient, "caos-cockroach")
		if err != nil {
			return nil, nil, err
		}
		queriers = append(queriers, queryNS)
		destroyers = append(destroyers, destroyNS)

		queryCert, destroyCert, err := certificate.AdaptFunc(k8sClient, "caos-cockroach", []string{"root"})
		if err != nil {
			return nil, nil, err
		}
		queriers = append(queriers, queryCert)
		destroyers = append(destroyers, destroyCert)

		return func() (cockroachdb.EnsureFunc, error) {
				ensurers := make([]cockroachdb.EnsureFunc, 0)
				for _, querier := range queriers {
					ensurer, err := querier()
					if err != nil {
						return nil, err
					}
					ensurers = append(ensurers, ensurer)
				}

				return func() error {
					for _, ensurer := range ensurers {
						if err := ensurer(); err != nil {
							return err
						}
					}
					return nil
				}, nil
			}, func() error {
				return nil
			},
			nil
	}
}
