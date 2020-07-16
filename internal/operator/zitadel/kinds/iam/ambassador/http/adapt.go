package http

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/ambassador/mapping"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/networking/core"
)

func AdaptFunc(
	namespace string,
	labels map[string]string,
	httpUrl string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {
	adminRName := "admin-rest-v1"
	mgmtName := "mgmt-v1"
	oauthName := "oauth-v1"
	authRName := "auth-rest-v1"
	authorizeName := "authorize-v1"
	endsessionName := "endsession-v1"
	issuerName := "issuer-v1"

	destroyAdminR, err := mapping.AdaptFuncToDestroy(adminRName, namespace)
	if err != nil {
		return nil, nil, err
	}

	destroyMgmtRest, err := mapping.AdaptFuncToDestroy(mgmtName, namespace)
	if err != nil {
		return nil, nil, err
	}

	destroyOAuthv2, err := mapping.AdaptFuncToDestroy(oauthName, namespace)
	if err != nil {
		return nil, nil, err
	}

	destroyAuthR, err := mapping.AdaptFuncToDestroy(authRName, namespace)
	if err != nil {
		return nil, nil, err
	}

	destroyAuthorize, err := mapping.AdaptFuncToDestroy(authorizeName, namespace)
	if err != nil {
		return nil, nil, err
	}

	destroyEndsession, err := mapping.AdaptFuncToDestroy(endsessionName, namespace)
	if err != nil {
		return nil, nil, err
	}

	destroyIssuer, err := mapping.AdaptFuncToDestroy(issuerName, namespace)
	if err != nil {
		return nil, nil, err
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

			currentNW, err := core.ParseQueriedForNetworking(queried)
			if err != nil {
				return nil, err
			}

			accountsDomain := currentNW.GetAccountsSubDomain() + "." + currentNW.GetDomain()
			apiDomain := currentNW.GetAPISubDomain() + "." + currentNW.GetDomain()
			issuerDomain := currentNW.GetIssuerSubDomain() + "." + currentNW.GetDomain()

			queryAdminR, err := mapping.AdaptFuncToEnsure(
				adminRName,
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
				return nil, err
			}

			queryMgmtRest, err := mapping.AdaptFuncToEnsure(
				mgmtName,
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
				return nil, err
			}

			queryOAuthv2, err := mapping.AdaptFuncToEnsure(
				oauthName,
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
				return nil, err
			}

			queryAuthR, err := mapping.AdaptFuncToEnsure(
				authRName,
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
				return nil, err
			}

			queryAuthorize, err := mapping.AdaptFuncToEnsure(
				authorizeName,
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
				return nil, err
			}

			queryEndsession, err := mapping.AdaptFuncToEnsure(
				endsessionName,
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
				return nil, err
			}

			queryIssuer, err := mapping.AdaptFuncToEnsure(
				issuerName,
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
				return nil, err
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

			return zitadel.QueriersToEnsureFunc(queriers, k8sClient, queried)
		},
		zitadel.DestroyersToDestroyFunc(destroyers),
		nil
}
