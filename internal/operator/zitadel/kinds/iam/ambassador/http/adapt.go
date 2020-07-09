package http

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/ambassador/mapping"
	"github.com/caos/orbos/internal/operator/zitadel"
)

func AdaptFunc(
	namespace string,
	labels map[string]string,
	httpUrl string,
	accountsDomain string,
	apiDomain string,
	issuerDomain string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {
	queryAdminR, destroyAdminR, err := mapping.AdaptFunc(
		"admin-rest-v1",
		namespace,
		labels,
		false,
		apiDomain,
		"/admin/v1",
		"",
		httpUrl,
		"30000",
		"30000",
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	queryMgmtRest, destroyMgmtRest, err := mapping.AdaptFunc(
		"mgmt-v1",
		namespace,
		labels,
		false,
		apiDomain,
		"/management/v1/",
		"",
		httpUrl,
		"30000",
		"30000",
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	queryOAuthv2, destroyOAuthv2, err := mapping.AdaptFunc(
		"oauth-v1",
		namespace,
		labels,
		false,
		apiDomain,
		"/oauth/v2/",
		"",
		httpUrl,
		"30000",
		"30000",
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	queryAuthR, destroyAuthR, err := mapping.AdaptFunc(
		"auth-rest-v1",
		namespace,
		labels,
		false,
		apiDomain,
		"/auth/v1/",
		"",
		httpUrl,
		"30000",
		"30000",
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	queryAuthorize, destroyAuthorize, err := mapping.AdaptFunc(
		"authorize-v1",
		namespace,
		labels,
		false,
		accountsDomain,
		"/oauth/v2/authorize",
		"",
		httpUrl,
		"30000",
		"30000",
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	queryEndsession, destroyEndsession, err := mapping.AdaptFunc(
		"endsession-v1",
		namespace,
		labels,
		false,
		accountsDomain,
		"/oauth/v2/endsession",
		"",
		httpUrl,
		"30000",
		"30000",
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	queryIssuer, destroyIssuer, err := mapping.AdaptFunc(
		"issuer-v1",
		namespace,
		labels,
		false,
		issuerDomain,
		"/.well-known/openid-configuration",
		"/oauth/v2/.well-known/openid-configuration",
		httpUrl,
		"30000",
		"30000",
		nil,
	)
	if err != nil {
		return nil, nil, err
	}

	queriers := []zitadel.QueryFunc{
		zitadel.ResourceQueryToZitadelQuery(queryAdminR),
		zitadel.ResourceQueryToZitadelQuery(queryMgmtRest),
		zitadel.ResourceQueryToZitadelQuery(queryOAuthv2),
		zitadel.ResourceQueryToZitadelQuery(queryAuthR),
		zitadel.ResourceQueryToZitadelQuery(queryAuthorize),
		zitadel.ResourceQueryToZitadelQuery(queryEndsession),
		zitadel.ResourceQueryToZitadelQuery(queryIssuer),
	}
	destroyers := []zitadel.DestroyFunc{
		zitadel.ResourceDestroyToZitadelDestroy(destroyAdminR),
		zitadel.ResourceDestroyToZitadelDestroy(destroyMgmtRest),
		zitadel.ResourceDestroyToZitadelDestroy(destroyOAuthv2),
		zitadel.ResourceDestroyToZitadelDestroy(destroyAuthR),
		zitadel.ResourceDestroyToZitadelDestroy(destroyAuthorize),
		zitadel.ResourceDestroyToZitadelDestroy(destroyEndsession),
		zitadel.ResourceDestroyToZitadelDestroy(destroyIssuer),
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
