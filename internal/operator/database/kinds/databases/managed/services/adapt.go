package services

import (
	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/mntr"
	kubernetes2 "github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources/service"
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
	core.QueryFunc,
	core.DestroyFunc,
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
	destroyers := []core.DestroyFunc{
		core.ResourceDestroyToZitadelDestroy(destroySPD),
		core.ResourceDestroyToZitadelDestroy(destroySP),
		core.ResourceDestroyToZitadelDestroy(destroyS),
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

	queriers := []core.QueryFunc{
		core.ResourceQueryToZitadelQuery(querySPD),
		core.ResourceQueryToZitadelQuery(querySP),
		core.ResourceQueryToZitadelQuery(queryS),
	}

	return func(k8sClient *kubernetes2.Client, queried map[string]interface{}) (core.EnsureFunc, error) {

			return core.QueriersToEnsureFunc(internalMonitor, false, queriers, k8sClient, queried)
		},
		core.DestroyersToDestroyFunc(internalMonitor, destroyers),
		nil

}
