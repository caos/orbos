package services

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/service"
	"github.com/caos/orbos/internal/operator/zitadel"
)

func AdaptFunc(
	namespace string,
	labels map[string]string,
	grpcServiceName string,
	grpcPort int,
	httpServiceName string,
	httpPort int,
	uiServiceName string,
	uiPort int,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {
	grpcPorts := []service.Port{
		{Name: "grpc", Port: grpcPort, TargetPort: "grpc"},
	}
	queryGRPC, destroyGRPC, err := service.AdaptFunc(grpcServiceName, namespace, labels, grpcPorts, "", labels, false, "", "")
	if err != nil {
		return nil, nil, err
	}

	httpPorts := []service.Port{
		{Name: "http", Port: httpPort, TargetPort: "http"},
	}
	queryHTTP, destroyHTTP, err := service.AdaptFunc(httpServiceName, namespace, labels, httpPorts, "", labels, false, "", "")
	if err != nil {
		return nil, nil, err
	}

	uiPorts := []service.Port{
		{Name: "ui", Port: uiPort, TargetPort: "ui"},
	}
	queryUI, destroyUI, err := service.AdaptFunc(uiServiceName, namespace, labels, uiPorts, "", labels, false, "", "")
	if err != nil {
		return nil, nil, err
	}

	queriers := make([]zitadel.QueryFunc, 0)
	queriers = []zitadel.QueryFunc{
		zitadel.ResourceQueryToZitadelQuery(queryGRPC),
		zitadel.ResourceQueryToZitadelQuery(queryHTTP),
		zitadel.ResourceQueryToZitadelQuery(queryUI),
	}

	destroyers := make([]zitadel.DestroyFunc, 0)
	destroyers = []zitadel.DestroyFunc{
		zitadel.ResourceDestroyToZitadelDestroy(destroyGRPC),
		zitadel.ResourceDestroyToZitadelDestroy(destroyHTTP),
		zitadel.ResourceDestroyToZitadelDestroy(destroyUI),
	}

	return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
			return zitadel.QueriersToEnsureFunc(queriers, k8sClient, queried)
		},
		zitadel.DestroyersToDestroyFunc(destroyers),
		nil
}
