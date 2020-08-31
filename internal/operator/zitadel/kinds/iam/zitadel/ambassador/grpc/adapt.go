package grpc

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/ambassador/mapping"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/ambassador/module"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/networking/core"
	"github.com/caos/orbos/mntr"
)

func AdaptFunc(
	monitor mntr.Monitor,
	namespace string,
	labels map[string]string,
	grpcURL string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {
	internalMonitor := monitor.WithField("part", "grpc")

	adminMName := "admin-grpc-v1"
	authMName := "auth-grpc-v1"
	mgmtMName := "mgmt-grpc-v1"
	moduleName := "ambassador"

	destroyAdminG, err := mapping.AdaptFuncToDestroy(namespace, adminMName)
	if err != nil {
		return nil, nil, err
	}
	destroyAuthG, err := mapping.AdaptFuncToDestroy(namespace, authMName)
	if err != nil {
		return nil, nil, err
	}
	destroyMgmtGRPC, err := mapping.AdaptFuncToDestroy(namespace, mgmtMName)
	if err != nil {
		return nil, nil, err
	}
	destroyModule, err := module.AdaptFuncToDestroy("caos-system", moduleName)
	if err != nil {
		return nil, nil, err
	}

	destroyers := []zitadel.DestroyFunc{
		zitadel.ResourceDestroyToZitadelDestroy(destroyAdminG),
		zitadel.ResourceDestroyToZitadelDestroy(destroyAuthG),
		zitadel.ResourceDestroyToZitadelDestroy(destroyMgmtGRPC),
		zitadel.ResourceDestroyToZitadelDestroy(destroyModule),
	}

	return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
			crd, err := k8sClient.CheckCRD("mappings.getambassador.io")
			if crd == nil || err != nil {
				return func(k8sClient *kubernetes.Client) error { return nil }, nil
			}

			crd, err = k8sClient.CheckCRD("modules.getambassador.io")
			if crd == nil || err != nil {
				return func(k8sClient *kubernetes.Client) error { return nil }, nil
			}

			currentNW, err := core.ParseQueriedForNetworking(queried)
			if err != nil {
				return nil, err
			}

			apiDomain := currentNW.GetAPISubDomain() + "." + currentNW.GetDomain()
			consoleDomain := currentNW.GetConsoleSubDomain() + "." + currentNW.GetDomain()
			_ = consoleDomain

			queryModule, err := module.AdaptFuncToEnsure("caos-system", moduleName, labels, &module.Config{EnableGrpcWeb: true})

			cors := &mapping.CORS{
				Origins:        "*",
				Methods:        "POST, GET, OPTIONS, DELETE, PUT",
				Headers:        "*",
				Credentials:    true,
				ExposedHeaders: "*",
				MaxAge:         "86400",
			}

			queryAdminG, err := mapping.AdaptFuncToEnsure(
				namespace,
				adminMName,
				labels,
				true,
				apiDomain,
				"/caos.zitadel.admin.api.v1.AdminService/",
				"",
				grpcURL,
				"30000",
				"30000",
				cors,
			)
			if err != nil {
				return nil, err
			}

			queryAuthG, err := mapping.AdaptFuncToEnsure(
				namespace,
				authMName,
				labels,
				true,
				apiDomain,
				"/caos.zitadel.auth.api.v1.AuthService/",
				"",
				grpcURL,
				"30000",
				"30000",
				cors,
			)
			if err != nil {
				return nil, err
			}

			queryMgmtGRPC, err := mapping.AdaptFuncToEnsure(
				namespace,
				mgmtMName,
				labels,
				true,
				apiDomain,
				"/caos.zitadel.management.api.v1.ManagementService/",
				"",
				grpcURL,
				"30000",
				"30000",
				cors,
			)
			if err != nil {
				return nil, err
			}

			queriers := []zitadel.QueryFunc{
				zitadel.ResourceQueryToZitadelQuery(queryModule),
				zitadel.ResourceQueryToZitadelQuery(queryAdminG),
				zitadel.ResourceQueryToZitadelQuery(queryAuthG),
				zitadel.ResourceQueryToZitadelQuery(queryMgmtGRPC),
			}

			return zitadel.QueriersToEnsureFunc(internalMonitor, false, queriers, k8sClient, queried)
		},
		zitadel.DestroyersToDestroyFunc(internalMonitor, destroyers),
		nil
}
