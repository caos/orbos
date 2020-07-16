package ui

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/ambassador/mapping"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/networking/core"
)

func AdaptFunc(
	namespace string,
	labels map[string]string,
	uiURL string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {

	consoleName := "console-v1"
	accountsName := "accounts-v1"

	destroyAcc, err := mapping.AdaptFuncToDestroy(accountsName, namespace)
	if err != nil {
		return nil, nil, err
	}

	destroyConsole, err := mapping.AdaptFuncToDestroy(consoleName, namespace)
	if err != nil {
		return nil, nil, err
	}

	destroyers := []zitadel.DestroyFunc{
		zitadel.ResourceDestroyToZitadelDestroy(destroyAcc),
		zitadel.ResourceDestroyToZitadelDestroy(destroyConsole),
	}

	return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
			crd, err := k8sClient.CheckCRD("mappings.getambassador.io")
			if crd == nil || err != nil {
				return func(k8sClient *kubernetes.Client) error { return nil }, nil
			}

			currentNW, err := core.ParseQueriedForNetworking(queried)
			if err != nil {
				return nil, err
			}

			accountsDomain := currentNW.GetAccountsSubDomain() + "." + currentNW.GetDomain()
			consoleDomain := currentNW.GetConsoleSubDomain() + "." + currentNW.GetDomain()

			queryConsole, err := mapping.AdaptFuncToEnsure(
				consoleName,
				namespace,
				labels,
				false,
				consoleDomain,
				"/",
				"/console/",
				uiURL,
				"",
				"",
				nil,
			)
			if err != nil {
				return nil, err
			}

			queryAcc, err := mapping.AdaptFuncToEnsure(
				accountsName,
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
				return nil, err
			}

			queriers := []zitadel.QueryFunc{
				zitadel.ResourceQueryToZitadelQuery(queryConsole),
				zitadel.ResourceQueryToZitadelQuery(queryAcc),
			}

			return zitadel.QueriersToEnsureFunc(queriers, k8sClient, queried)
		},
		zitadel.DestroyersToDestroyFunc(destroyers),
		nil
}
