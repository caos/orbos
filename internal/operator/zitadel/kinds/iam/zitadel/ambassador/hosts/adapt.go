package hosts

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/ambassador/host"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/networking/core"
	"github.com/caos/orbos/mntr"
)

func AdaptFunc(
	monitor mntr.Monitor,
	namespace string,
	labels map[string]string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {
	internalMonitor := monitor.WithField("part", "hosts")

	accountsHostName := "accounts"
	apiHostName := "api"
	consoleHostName := "console"
	issuerHostName := "issuer"

	destroyAccounts, err := host.AdaptFuncToDestroy(namespace, accountsHostName)
	if err != nil {
		return nil, nil, err
	}

	destroyAPI, err := host.AdaptFuncToDestroy(namespace, apiHostName)
	if err != nil {
		return nil, nil, err
	}

	destroyConsole, err := host.AdaptFuncToDestroy(namespace, consoleHostName)
	if err != nil {
		return nil, nil, err
	}

	destroyIssuer, err := host.AdaptFuncToDestroy(namespace, issuerHostName)
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
			originCASecretName := currentNW.GetTlsCertName()

			accountsSelector := map[string]string{
				"hostname": accountsDomain,
			}
			queryAccounts, err := host.AdaptFuncToEnsure(namespace, accountsHostName, labels, accountsDomain, "none", "", accountsSelector, originCASecretName)
			if err != nil {
				return nil, err
			}

			apiSelector := map[string]string{
				"hostname": apiDomain,
			}
			queryAPI, err := host.AdaptFuncToEnsure(namespace, apiHostName, labels, apiDomain, "none", "", apiSelector, originCASecretName)
			if err != nil {
				return nil, err
			}

			consoleSelector := map[string]string{
				"hostname": consoleDomain,
			}
			queryConsole, err := host.AdaptFuncToEnsure(namespace, consoleHostName, labels, consoleDomain, "none", "", consoleSelector, originCASecretName)
			if err != nil {
				return nil, err
			}

			issuerSelector := map[string]string{
				"hostname": issuerDomain,
			}
			queryIssuer, err := host.AdaptFuncToEnsure(namespace, issuerHostName, labels, issuerDomain, "none", "", issuerSelector, originCASecretName)
			if err != nil {
				return nil, err
			}

			queriers := []zitadel.QueryFunc{
				zitadel.EnsureFuncToQueryFunc(currentNW.GetReadyCertificate()),
				zitadel.ResourceQueryToZitadelQuery(queryAccounts),
				zitadel.ResourceQueryToZitadelQuery(queryAPI),
				zitadel.ResourceQueryToZitadelQuery(queryConsole),
				zitadel.ResourceQueryToZitadelQuery(queryIssuer),
			}

			return zitadel.QueriersToEnsureFunc(internalMonitor, false, queriers, k8sClient, queried)
		},
		zitadel.DestroyersToDestroyFunc(internalMonitor, destroyers),
		nil
}
