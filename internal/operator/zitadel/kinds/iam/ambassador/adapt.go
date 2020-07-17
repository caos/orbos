package ambassador

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/ambassador/grpc"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/ambassador/hosts"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/ambassador/http"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/ambassador/ui"
	"github.com/caos/orbos/mntr"
)

func AdaptFunc(
	monitor mntr.Monitor,
	namespace string,
	labels map[string]string,
	grpcURL string,
	httpURL string,
	uiURL string,
	originCASecretName string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {
	internalMonitor := monitor.WithField("component", "ambassador")

	internalLabels := make(map[string]string, 0)
	for k, v := range labels {
		internalLabels[k] = v
	}
	internalLabels["app.kubernetes.io/component"] = "ambassador"

	queryHosts, destroyHosts, err := hosts.AdaptFunc(internalMonitor, namespace, labels, originCASecretName)
	if err != nil {
		return nil, nil, err
	}

	queryGRPC, destroyGRPC, err := grpc.AdaptFunc(internalMonitor, namespace, labels, grpcURL)
	if err != nil {
		return nil, nil, err
	}

	queryUI, destroyHTTP, err := ui.AdaptFunc(internalMonitor, namespace, labels, uiURL)
	if err != nil {
		return nil, nil, err
	}

	queryHTTP, destroyUI, err := http.AdaptFunc(internalMonitor, namespace, labels, httpURL)
	if err != nil {
		return nil, nil, err
	}

	queriers := []zitadel.QueryFunc{
		queryHosts,
		queryGRPC,
		queryUI,
		queryHTTP,
	}
	destroyers := []zitadel.DestroyFunc{
		destroyGRPC,
		destroyHTTP,
		destroyUI,
		destroyHosts,
	}
	return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
			return zitadel.QueriersToEnsureFunc(internalMonitor, true, queriers, k8sClient, queried)
		},
		zitadel.DestroyersToDestroyFunc(internalMonitor, destroyers),
		nil
}
