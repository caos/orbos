package ui

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/ambassador/mapping"
	"github.com/caos/orbos/internal/operator/zitadel"
)

func AdaptFunc(
	namespace string,
	labels map[string]string,
	uiURL string,
	accountsDomain string,
	consoleDomain string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {

	queriers := []zitadel.QueryFunc{}
	destroyers := []zitadel.DestroyFunc{}

	if accountsDomain != "" {
		queryAcc, destroyAcc, err := mapping.AdaptFunc(
			"accounts-v1",
			namespace,
			labels,
			false,
			accountsDomain,
			"/",
			"/login/",
			uiURL,
			"30000",
			"30000",
			nil,
		)
		if err != nil {
			return nil, nil, err
		}
		queriers = append(queriers, zitadel.ResourceQueryToZitadelQuery(queryAcc))
		destroyers = append(destroyers, zitadel.ResourceDestroyToZitadelDestroy(destroyAcc))
	}

	if consoleDomain != "" {
		queryConsole, destroyConsole, err := mapping.AdaptFunc(
			"console-v1",
			namespace,
			labels,
			false,
			"console.zitadel.dev",
			"/",
			"/console/",
			uiURL,
			"",
			"",
			nil,
		)
		if err != nil {
			return nil, nil, err
		}
		queriers = append(queriers, zitadel.ResourceQueryToZitadelQuery(queryConsole))
		destroyers = append(destroyers, zitadel.ResourceDestroyToZitadelDestroy(destroyConsole))
	}

	return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
			crd, err := k8sClient.CheckCRD("mappings.getambassador.io")
			if crd == nil || err != nil {
				return func(k8sClient *kubernetes.Client) error { return nil }, nil
			}

			return zitadel.QueriersToEnsureFunc(queriers, k8sClient, queried)
		},
		zitadel.DestroyersToDestroyFunc(destroyers),
		nil
}
