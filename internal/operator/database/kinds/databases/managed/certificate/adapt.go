package certificate

import (
	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/managed/certificate/client"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/managed/certificate/node"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
)

func AdaptFunc(
	monitor mntr.Monitor,
	namespace string,
	labels map[string]string,
	clusterDns string,
) (
	core.QueryFunc,
	core.DestroyFunc,
	func(user string) (core.QueryFunc, error),
	func(user string) (core.DestroyFunc, error),
	func(k8sClient *kubernetes.Client) ([]string, error),
	error,
) {
	cMonitor := monitor.WithField("component", "certificates")

	queryNode, destroyNode, err := node.AdaptFunc(
		cMonitor,
		namespace,
		labels,
		clusterDns,
	)
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	queriers := []core.QueryFunc{
		queryNode,
	}

	destroyers := []core.DestroyFunc{
		destroyNode,
	}

	return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (core.EnsureFunc, error) {
			return core.QueriersToEnsureFunc(cMonitor, false, queriers, k8sClient, queried)
		},
		core.DestroyersToDestroyFunc(cMonitor, destroyers),
		func(user string) (core.QueryFunc, error) {
			query, _, err := client.AdaptFunc(
				cMonitor,
				namespace,
				labels,
			)
			if err != nil {
				return nil, err
			}
			queryClient := query(user)

			return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (core.EnsureFunc, error) {
				_, err := queryNode(k8sClient, queried)
				if err != nil {
					return nil, err
				}

				return queryClient(k8sClient, queried)
			}, nil
		},
		func(user string) (core.DestroyFunc, error) {
			_, destroy, err := client.AdaptFunc(
				cMonitor,
				namespace,
				labels,
			)
			if err != nil {
				return nil, err
			}

			return destroy(user), nil
		},
		func(k8sClient *kubernetes.Client) ([]string, error) {
			return client.QueryCertificates(namespace, labels, k8sClient)
		},
		nil
}
