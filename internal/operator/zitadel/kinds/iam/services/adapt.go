package services

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/service"
	"github.com/caos/orbos/internal/operator/zitadel"
)

func AdaptFunc(
	namespace string,
	labels map[string]string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {

	accountsPorts := []service.Port{
		{Name: "http", Port: 80, TargetPort: "accounts-http"},
	}
	queryS, destroyS, err := service.AdaptFunc("accounts-v1", namespace, labels, accountsPorts, "", labels, false, "", "")
	if err != nil {
		return nil, nil, err
	}

	serviceApiAdminPorts := []service.Port{
		{Name: "rest", Port: 80, TargetPort: "admin-rest"},
		{Name: "grpc", Port: 8080, TargetPort: "admin-grpc"},
	}
	querySAA, destroySAA, err := service.AdaptFunc("api-admin-v1", namespace, labels, serviceApiAdminPorts, "", labels, false, "", "")
	if err != nil {
		return nil, nil, err
	}

	serviceApiAuthPorts := []service.Port{
		{Name: "rest", Port: 80, TargetPort: "auth-rest"},
		{Name: "issuer", Port: 7070, TargetPort: "issuer-rest"},
		{Name: "grpc", Port: 8080, TargetPort: "auth-grpc"},
	}
	queryAA, destroyAA, err := service.AdaptFunc("api-auth-v1", namespace, labels, serviceApiAuthPorts, "", labels, false, "", "")
	if err != nil {
		return nil, nil, err
	}

	serviceApiMgmtPorts := []service.Port{
		{Name: "rest", Port: 80, TargetPort: "management-rest"},
		{Name: "grpc", Port: 8080, TargetPort: "management-grpc"},
	}
	querySAM, destroySAM, err := service.AdaptFunc("api-management-v1", namespace, labels, serviceApiMgmtPorts, "", labels, false, "", "")
	if err != nil {
		return nil, nil, err
	}

	serviceConsolePorts := []service.Port{
		{Name: "http", Port: 80, TargetPort: "console-http"},
	}
	querySC, destroySC, err := service.AdaptFunc("console-v1", namespace, labels, serviceConsolePorts, "", labels, false, "", "")
	if err != nil {
		return nil, nil, err
	}

	queriers := make([]zitadel.QueryFunc, 0)
	queriers = []zitadel.QueryFunc{
		zitadel.ResourceQueryToZitadelQuery(queryS),
		zitadel.ResourceQueryToZitadelQuery(querySAA),
		zitadel.ResourceQueryToZitadelQuery(queryAA),
		zitadel.ResourceQueryToZitadelQuery(querySAM),
		zitadel.ResourceQueryToZitadelQuery(querySC),
	}

	destroyers := make([]zitadel.DestroyFunc, 0)
	destroyers = []zitadel.DestroyFunc{
		zitadel.ResourceDestroyToZitadelDestroy(destroyS),
		zitadel.ResourceDestroyToZitadelDestroy(destroySAA),
		zitadel.ResourceDestroyToZitadelDestroy(destroyAA),
		zitadel.ResourceDestroyToZitadelDestroy(destroySAM),
		zitadel.ResourceDestroyToZitadelDestroy(destroySC),
	}

	return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
			return zitadel.QueriersToEnsureFunc(queriers, k8sClient, queried)
		},
		zitadel.DestroyersToDestroyFunc(destroyers),
		nil
}
