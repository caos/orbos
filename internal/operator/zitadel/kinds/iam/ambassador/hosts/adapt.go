package hosts

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/ambassador/host"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/networking/core"
)

func AdaptFunc(
	namespace string,
	labels map[string]string,
	originCASecretName string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {
	accountsHostName := "accounts"
	apiHostName := "api"
	consoleHostName := "console"
	issuerHostName := "issuer"

	destroyAccounts, err := host.AdaptFuncToDestroy(accountsHostName, namespace)
	if err != nil {
		return nil, nil, err
	}

	destroyAPI, err := host.AdaptFuncToDestroy(apiHostName, namespace)
	if err != nil {
		return nil, nil, err
	}

	destroyConsole, err := host.AdaptFuncToDestroy(consoleHostName, namespace)
	if err != nil {
		return nil, nil, err
	}

	destroyIssuer, err := host.AdaptFuncToDestroy(issuerHostName, namespace)
	if err != nil {
		return nil, nil, err
	}

	destroyers := []zitadel.DestroyFunc{
		zitadel.ResourceDestroyToZitadelDestroy(destroyAccounts),
		zitadel.ResourceDestroyToZitadelDestroy(destroyAPI),
		zitadel.ResourceDestroyToZitadelDestroy(destroyConsole),
		zitadel.ResourceDestroyToZitadelDestroy(destroyIssuer),
	}

	return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
			crd, err := k8sClient.CheckCRD("hosts.getambassador.io")
			if crd == nil || err != nil {
				return func(k8sClient *kubernetes.Client) error { return nil }, nil
			}

			currentNW, err := core.ParseQueriedForNetworking(queried)
			if err != nil {
				return nil, err
			}

			accountsDomain := currentNW.GetAccountsSubDomain() + "." + currentNW.GetDomain()
			apiDomain := currentNW.GetAPISubDomain() + "." + currentNW.GetDomain()
			consoleDomain := currentNW.GetConsoleSubDomain() + "." + currentNW.GetDomain()
			issuerDomain := currentNW.GetIssuerSubDomain() + "." + currentNW.GetDomain()

			accountsSelector := map[string]string{
				"hostname": accountsDomain,
			}
			queryAccounts, err := host.AdaptFuncToEnsure(accountsHostName, namespace, labels, accountsDomain, "none", "", accountsSelector, originCASecretName)
			if err != nil {
				return nil, err
			}

			apiSelector := map[string]string{
				"hostname": apiDomain,
			}
			queryAPI, err := host.AdaptFuncToEnsure(apiHostName, namespace, labels, apiDomain, "none", "", apiSelector, originCASecretName)
			if err != nil {
				return nil, err
			}

			consoleSelector := map[string]string{
				"hostname": consoleDomain,
			}
			queryConsole, err := host.AdaptFuncToEnsure(consoleHostName, namespace, labels, consoleDomain, "none", "", consoleSelector, originCASecretName)
			if err != nil {
				return nil, err
			}

			issuerSelector := map[string]string{
				"hostname": issuerDomain,
			}
			queryIssuer, err := host.AdaptFuncToEnsure(issuerHostName, namespace, labels, issuerDomain, "none", "", issuerSelector, originCASecretName)
			if err != nil {
				return nil, err
			}

			queriers := []zitadel.QueryFunc{
				zitadel.ResourceQueryToZitadelQuery(queryAccounts),
				zitadel.ResourceQueryToZitadelQuery(queryAPI),
				zitadel.ResourceQueryToZitadelQuery(queryConsole),
				zitadel.ResourceQueryToZitadelQuery(queryIssuer),
			}

			return zitadel.QueriersToEnsureFunc(queriers, k8sClient, queried)
		},
		zitadel.DestroyersToDestroyFunc(destroyers),
		nil
}
