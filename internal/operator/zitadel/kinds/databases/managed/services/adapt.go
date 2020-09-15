package services

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/service"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/mntr"
	"strconv"
)

func AdaptFunc(
	monitor mntr.Monitor,
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
	internalMonitor := monitor.WithField("component", "services")

	destroySPD, err := service.AdaptFuncToDestroy("default", publicServiceName)
	if err != nil {
		return nil, nil, err
	}
	destroySP, err := service.AdaptFuncToDestroy(namespace, publicServiceName)
	if err != nil {
		return nil, nil, err
	}
	destroyS, err := service.AdaptFuncToDestroy(namespace, serviceName)
	if err != nil {
		return nil, nil, err
	}
	destroyers := []zitadel.DestroyFunc{
		zitadel.ResourceDestroyToZitadelDestroy(destroySPD),
		zitadel.ResourceDestroyToZitadelDestroy(destroySP),
		zitadel.ResourceDestroyToZitadelDestroy(destroyS),
	}

	publicLabels := map[string]string{}
	for k, v := range labels {
		publicLabels[k] = v
	}
	publicLabels["zitadel.caos.ch/servicetype"] = "public"

	internalLabels := map[string]string{}
	for k, v := range labels {
		internalLabels[k] = v
	}
	internalLabels["zitadel.caos.ch/servicetype"] = "internal"

	ports := []service.Port{
		{Port: 26257, TargetPort: strconv.Itoa(int(cockroachPort)), Name: "grpc"},
		{Port: 8080, TargetPort: strconv.Itoa(int(cockroachHTTPPort)), Name: "http"},
	}
	querySPD, err := service.AdaptFuncToEnsure("default", publicServiceName, publicLabels, ports, "", labels, false, "", "")
	if err != nil {
		return nil, nil, err
	}
	querySP, err := service.AdaptFuncToEnsure(namespace, publicServiceName, publicLabels, ports, "", labels, false, "", "")
	if err != nil {
		return nil, nil, err
	}
	queryS, err := service.AdaptFuncToEnsure(namespace, serviceName, internalLabels, ports, "", labels, true, "None", "")
	if err != nil {
		return nil, nil, err
	}

	queriers := []zitadel.QueryFunc{
		zitadel.ResourceQueryToZitadelQuery(querySPD),
		zitadel.ResourceQueryToZitadelQuery(querySP),
		zitadel.ResourceQueryToZitadelQuery(queryS),
	}

	return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {

			return zitadel.QueriersToEnsureFunc(internalMonitor, false, queriers, k8sClient, queried)
		},
		zitadel.DestroyersToDestroyFunc(internalMonitor, destroyers),
		nil

}
