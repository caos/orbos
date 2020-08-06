package http

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/ambassador/mapping"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/networking/core"
	"github.com/caos/orbos/mntr"
)

func AdaptFunc(
	monitor mntr.Monitor,
	namespace string,
	labels map[string]string,
	httpUrl string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {
	internalMonitor := monitor.WithField("part", "http")

	adminRName := "admin-rest-v1"
	mgmtName := "mgmt-v1"
	oauthName := "oauth-v1"
	authRName := "auth-rest-v1"
	authorizeName := "authorize-v1"
	endsessionName := "endsession-v1"
	issuerName := "issuer-v1"

	destroyAdminR, err := mapping.AdaptFuncToDestroy(namespace, adminRName)
	if err != nil {
		return nil, nil, err
	}

	destroyMgmtRest, err := mapping.AdaptFuncToDestroy(namespace, mgmtName)
	if err != nil {
		return nil, nil, err
	}

	destroyOAuthv2, err := mapping.AdaptFuncToDestroy(namespace, oauthName)
	if err != nil {
		return nil, nil, err
	}

	destroyAuthR, err := mapping.AdaptFuncToDestroy(namespace, authRName)
	if err != nil {
		return nil, nil, err
	}

	destroyAuthorize, err := mapping.AdaptFuncToDestroy(namespace, authorizeName)
	if err != nil {
		return nil, nil, err
	}

	destroyEndsession, err := mapping.AdaptFuncToDestroy(namespace, endsessionName)
	if err != nil {
		return nil, nil, err
	}

	destroyIssuer, err := mapping.AdaptFuncToDestroy(namespace, issuerName)
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
				namespace,
				adminRName,
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
				namespace,
				mgmtName,
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
				namespace,
				oauthName,
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
				namespace,
				authRName,
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
				namespace,
				authorizeName,
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
				namespace,
				endsessionName,
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
				namespace,
				issuerName,
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

			return zitadel.QueriersToEnsureFunc(internalMonitor, false, queriers, k8sClient, queried)
		},
		zitadel.DestroyersToDestroyFunc(internalMonitor, destroyers),
		nil
}
