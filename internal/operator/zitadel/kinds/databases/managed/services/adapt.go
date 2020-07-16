package services

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/service"
	"github.com/caos/orbos/internal/operator/zitadel"
	"strconv"
)

func AdaptFunc(
	namespace string,
	publicServiceName string,
	serviceName string,
	labels map[string]string,
	cockroachPort int32,
	cockroachHTTPPort int32,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {
	destroySPD, err := service.AdaptFuncToDestroy(publicServiceName, "default")
	if err != nil {
		return nil, nil, err
	}
	destroySP, err := service.AdaptFuncToDestroy(publicServiceName, namespace)
	if err != nil {
		return nil, nil, err
	}
	destroyS, err := service.AdaptFuncToDestroy(serviceName, namespace)
	if err != nil {
		return nil, nil, err
	}
	destroyers := []zitadel.DestroyFunc{
		zitadel.ResourceDestroyToZitadelDestroy(destroySPD),
		zitadel.ResourceDestroyToZitadelDestroy(destroySP),
		zitadel.ResourceDestroyToZitadelDestroy(destroyS),
	}

	ports := []service.Port{
		{Port: 26257, TargetPort: strconv.Itoa(int(cockroachPort)), Name: "grpc"},
		{Port: 8080, TargetPort: strconv.Itoa(int(cockroachHTTPPort)), Name: "http"},
	}
	querySPD, err := service.AdaptFuncToEnsure(publicServiceName, "default", labels, ports, "", labels, false, "", "")
	if err != nil {
		return nil, nil, err
	}
	querySP, err := service.AdaptFuncToEnsure(publicServiceName, namespace, labels, ports, "", labels, false, "", "")
	if err != nil {
		return nil, nil, err
	}
	queryS, err := service.AdaptFuncToEnsure(serviceName, namespace, labels, ports, "", labels, true, "None", "")
	if err != nil {
		return nil, nil, err
	}

	queriers := []zitadel.QueryFunc{
		zitadel.ResourceQueryToZitadelQuery(querySPD),
		zitadel.ResourceQueryToZitadelQuery(querySP),
		zitadel.ResourceQueryToZitadelQuery(queryS),
	}

	return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
			return zitadel.QueriersToEnsureFunc(queriers, k8sClient, queried)
		},
		zitadel.DestroyersToDestroyFunc(destroyers),
		nil

}
