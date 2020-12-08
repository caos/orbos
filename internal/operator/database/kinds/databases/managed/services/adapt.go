package services

import (
	"strconv"

	"github.com/caos/orbos/pkg/labels"

	"github.com/caos/orbos/internal/operator/core"
	"github.com/caos/orbos/mntr"
	"github.com/caos/orbos/pkg/kubernetes"
	"github.com/caos/orbos/pkg/kubernetes/resources/service"
)

func AdaptFunc(
	monitor mntr.Monitor,
	namespace string,
	publicServiceNameLabels *labels.Name,
	serviceNameLabels *labels.Name,
	cockroachLabels *labels.Name,
	cockroachPort int32,
	cockroachHTTPPort int32,
) (
	core.QueryFunc,
	core.DestroyFunc,
	error,
) {
	internalMonitor := monitor.WithField("type", "services")

	destroySPD, err := service.AdaptFuncToDestroy("default", publicServiceNameLabels.Name())
	if err != nil {
		return nil, nil, err
	}
	destroySP, err := service.AdaptFuncToDestroy(namespace, publicServiceNameLabels.Name())
	if err != nil {
		return nil, nil, err
	}
	destroyS, err := service.AdaptFuncToDestroy(namespace, serviceNameLabels.Name())
	if err != nil {
		return nil, nil, err
	}
	destroyers := []core.DestroyFunc{
		core.ResourceDestroyToZitadelDestroy(destroySPD),
		core.ResourceDestroyToZitadelDestroy(destroySP),
		core.ResourceDestroyToZitadelDestroy(destroyS),
	}

	ports := []service.Port{
		{Port: 26257, TargetPort: strconv.Itoa(int(cockroachPort)), Name: "grpc"},
		{Port: 8080, TargetPort: strconv.Itoa(int(cockroachHTTPPort)), Name: "http"},
	}
	querySPD, err := service.AdaptFuncToEnsure("default", publicServiceNameLabels, ports, "", cockroachLabels, false, "", "")
	if err != nil {
		return nil, nil, err
	}
	querySP, err := service.AdaptFuncToEnsure(namespace, publicServiceNameLabels, ports, "", cockroachLabels, false, "", "")
	if err != nil {
		return nil, nil, err
	}
	queryS, err := service.AdaptFuncToEnsure(namespace, serviceNameLabels, ports, "", cockroachLabels, true, "None", "")
	if err != nil {
		return nil, nil, err
	}

	queriers := []core.QueryFunc{
		core.ResourceQueryToZitadelQuery(querySPD),
		core.ResourceQueryToZitadelQuery(querySP),
		core.ResourceQueryToZitadelQuery(queryS),
	}

	return func(k8sClient kubernetes.ClientInt, queried map[string]interface{}) (core.EnsureFunc, error) {

			return core.QueriersToEnsureFunc(internalMonitor, false, queriers, k8sClient, queried)
		},
		core.DestroyersToDestroyFunc(internalMonitor, destroyers),
		nil

}
