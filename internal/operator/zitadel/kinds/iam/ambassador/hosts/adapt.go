package hosts

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/ambassador/host"
	"github.com/caos/orbos/internal/operator/zitadel"
)

func AdaptFunc(
	namespace string,
	labels map[string]string,
	accountsDomain string,
	apiDomain string,
	consoleDomain string,
	issuerDomain string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {
	accountsSelector := map[string]string{
		"hostname": accountsDomain,
	}
	queryAccounts, destroyAccounts, err := host.AdaptFunc("accounts", namespace, labels, accountsDomain, "none", "", accountsSelector, "tls-cert-wildcard")
	if err != nil {
		return nil, nil, err
	}

	apiSelector := map[string]string{
		"hostname": apiDomain,
	}
	queryAPI, destroyAPI, err := host.AdaptFunc("api", namespace, labels, apiDomain, "none", "", apiSelector, "tls-cert-wildcard")
	if err != nil {
		return nil, nil, err
	}

	consoleSelector := map[string]string{
		"hostname": consoleDomain,
	}
	queryConsole, destroyConsole, err := host.AdaptFunc("console", namespace, labels, consoleDomain, "none", "", consoleSelector, "tls-cert-wildcard")
	if err != nil {
		return nil, nil, err
	}

	issuerSelector := map[string]string{
		"hostname": issuerDomain,
	}
	queryIssuer, destroyIssuer, err := host.AdaptFunc("issuer", namespace, labels, issuerDomain, "none", "", issuerSelector, "tls-cert-wildcard")
	if err != nil {
		return nil, nil, err
	}

	queriers := []zitadel.QueryFunc{
		zitadel.ResourceQueryToZitadelQuery(queryAccounts),
		zitadel.ResourceQueryToZitadelQuery(queryAPI),
		zitadel.ResourceQueryToZitadelQuery(queryConsole),
		zitadel.ResourceQueryToZitadelQuery(queryIssuer),
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

			return zitadel.QueriersToEnsureFunc(queriers, k8sClient, queried)
		},
		zitadel.DestroyersToDestroyFunc(destroyers),
		nil
}
