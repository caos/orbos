package ambassador

import (
	"errors"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/ambassador/grpc"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/ambassador/hosts"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/ambassador/http"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/iam/ambassador/ui"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/networking/core"
	"github.com/caos/orbos/internal/tree"
)

func AdaptFunc(
	namespace string,
	labels map[string]string,
	grpcURL string,
	httpURL string,
	uiURL string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {

	internalLabels := make(map[string]string, 0)
	for k, v := range labels {
		internalLabels[k] = v
	}
	internalLabels["app.kubernetes.io/component"] = "ambassador"

	_, destroyGRPC, err := grpc.AdaptFunc(namespace, labels, grpcURL, "", "")
	if err != nil {
		return nil, nil, err
	}

	_, destroyHTTP, err := http.AdaptFunc(namespace, labels, httpURL, "", "", "")
	if err != nil {
		return nil, nil, err
	}

	_, destroyUI, err := ui.AdaptFunc(namespace, labels, uiURL, "", "")
	if err != nil {
		return nil, nil, err
	}

	_, destroyHosts, err := hosts.AdaptFunc(namespace, labels, "", "", "", "")
	if err != nil {
		return nil, nil, err
	}

	queriers := []zitadel.QueryFunc{}
	destroyers := []zitadel.DestroyFunc{
		destroyGRPC,
		destroyHTTP,
		destroyUI,
		destroyHosts,
	}
	return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
			crd, err := k8sClient.CheckCRD("hosts.getambassador.io")
			if crd == nil || err != nil {
				return func(k8sClient *kubernetes.Client) error { return nil }, nil
			}

			queriedNW, ok := queried["networking"]
			if !ok {
				return nil, errors.New("no current state for networking found")
			}
			current, ok := queriedNW.(*tree.Tree)
			if !ok {
				return nil, errors.New("current state does not fullfil interface")
			}
			currentNW, ok := current.Parsed.(core.NetworkingCurrent)
			if !ok {
				return nil, errors.New("current state does not fullfil interface")
			}

			accountsDomain := currentNW.GetAccountsSubDomain() + "." + currentNW.GetDomain()
			apiDomain := currentNW.GetAPISubDomain() + "." + currentNW.GetDomain()
			consoleDomain := currentNW.GetConsoleSubDomain() + "." + currentNW.GetDomain()
			issuerDomain := currentNW.GetIssuerSubDomain() + "." + currentNW.GetDomain()

			queryHosts, _, err := hosts.AdaptFunc(namespace, labels, accountsDomain, apiDomain, consoleDomain, issuerDomain)
			if err != nil {
				return nil, err
			}
			queriers = append(queriers, queryHosts)

			queryGRPC, _, err := grpc.AdaptFunc(namespace, labels, grpcURL, apiDomain, consoleDomain)
			if err != nil {
				return nil, err
			}
			queriers = append(queriers, queryGRPC)

			queryUI, _, err := ui.AdaptFunc(namespace, labels, uiURL, accountsDomain, consoleDomain)
			if err != nil {
				return nil, err
			}
			queriers = append(queriers, queryUI)

			queryHTTP, _, err := http.AdaptFunc(namespace, labels, httpURL, accountsDomain, apiDomain, issuerDomain)
			if err != nil {
				return nil, err
			}
			queriers = append(queriers, queryHTTP)

			return zitadel.QueriersToEnsureFunc(queriers, k8sClient, queried)
		},
		zitadel.DestroyersToDestroyFunc(destroyers),
		nil
}
