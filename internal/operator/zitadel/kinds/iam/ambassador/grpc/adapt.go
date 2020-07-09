package grpc

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/ambassador/mapping"
	"github.com/caos/orbos/internal/operator/zitadel"
)

func AdaptFunc(
	namespace string,
	labels map[string]string,
	grpcURL string,
	apiDomain string,
	consoleDomain string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {

	queryAdminG, destroyAdminG, err := mapping.AdaptFunc(
		"admin-grpc-v1",
		namespace,
		labels,
		true,
		apiDomain,
		"/caos.zitadel.admin.api.v1.AdminService/",
		"/login/",
		grpcURL,
		"30000",
		"30000",
		&mapping.CORS{
			Origins:        "https://" + consoleDomain,
			Methods:        "POST, GET, OPTIONS, DELETE, PUT",
			Headers:        "'*'",
			Credentials:    true,
			ExposedHeaders: "'*'",
			MaxAge:         "86400",
		},
	)
	if err != nil {
		return nil, nil, err
	}

	queryAuthG, destroyAuthG, err := mapping.AdaptFunc(
		"auth-grpc-v1",
		namespace,
		labels,
		true,
		apiDomain,
		"/caos.zitadel.auth.api.v1.AuthService/",
		"",
		grpcURL,
		"30000",
		"30000",
		&mapping.CORS{
			Origins:        "https://" + consoleDomain,
			Methods:        "POST, GET, OPTIONS, DELETE, PUT",
			Headers:        "'*'",
			Credentials:    true,
			ExposedHeaders: "'*'",
			MaxAge:         "86400",
		},
	)
	if err != nil {
		return nil, nil, err
	}

	queryMgmtGRPC, destroyMgmtGRPC, err := mapping.AdaptFunc(
		"mgmt-grpc-v1",
		namespace,
		labels,
		true,
		apiDomain,
		"/caos.zitadel.management.api.v1.ManagementService/",
		"",
		grpcURL,
		"30000",
		"30000",
		&mapping.CORS{
			Origins:        "https://" + consoleDomain,
			Methods:        "POST, GET, OPTIONS, DELETE, PUT",
			Headers:        "'*'",
			Credentials:    true,
			ExposedHeaders: "'*'",
			MaxAge:         "86400",
		},
	)
	if err != nil {
		return nil, nil, err
	}

	queriers := []zitadel.QueryFunc{
		zitadel.ResourceQueryToZitadelQuery(queryAdminG),
		zitadel.ResourceQueryToZitadelQuery(queryAuthG),
		zitadel.ResourceQueryToZitadelQuery(queryMgmtGRPC),
	}
	destroyers := []zitadel.DestroyFunc{
		zitadel.ResourceDestroyToZitadelDestroy(destroyAdminG),
		zitadel.ResourceDestroyToZitadelDestroy(destroyAuthG),
		zitadel.ResourceDestroyToZitadelDestroy(destroyMgmtGRPC),
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
