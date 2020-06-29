package iam

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/service"
	"github.com/caos/orbos/internal/operator/zitadel"
	"github.com/caos/orbos/internal/operator/zitadel/iam/deployment"
	"github.com/caos/orbos/internal/tree"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
	"io/ioutil"
)

func AdaptFunc() zitadel.AdaptFunc {
	return func(
		monitor mntr.Monitor,
		desired *tree.Tree,
		current *tree.Tree,
	) (
		zitadel.QueryFunc,
		zitadel.DestroyFunc,
		error,
	) {

		desiredKind, err := parseDesiredV0(desired)
		if err != nil {
			return nil, nil, errors.Wrap(err, "parsing desired state failed")
		}
		desired.Parsed = desiredKind

		data, err := ioutil.ReadFile("/Users/benz/.kube/config")
		dummyKubeconfig := string(data)
		k8sClient := kubernetes.NewK8sClient(monitor, &dummyKubeconfig)
		//if err := k8sClient.RefreshLocal(); err != nil {
		//	return nil, nil, err
		//}

		if !k8sClient.Available() {
			return nil, nil, errors.New("kubeconfig failed")
		}
		queriers := make([]resources.QueryFunc, 0)
		destroyers := make([]resources.DestroyFunc, 0)

		namespaceStr := "caos-zitadel"
		labels := map[string]string{"app.kubernetes.io/managed-by": "zitadel.caos.ch"}

		queryD, destroyD, err := deployment.AdaptFunc(k8sClient, namespaceStr, labels, desiredKind.Spec.ReplicaCount)
		if err != nil {
			return nil, nil, err
		}

		accountsPorts := []service.Port{
			{Name: "http", Port: 80, TargetPort: "accounts-http"},
		}
		queryS, destroyS, err := service.AdaptFunc(k8sClient, "accounts-v1", namespaceStr, labels, accountsPorts, "", labels, false, "", "")
		if err != nil {
			return nil, nil, err
		}

		serviceApiAdminPorts := []service.Port{
			{Name: "rest", Port: 80, TargetPort: "admin-rest"},
			{Name: "grpc", Port: 8080, TargetPort: "admin-grpc"},
		}
		querySAA, destroySAA, err := service.AdaptFunc(k8sClient, "api-admin-v1", namespaceStr, labels, serviceApiAdminPorts, "", labels, false, "", "")
		if err != nil {
			return nil, nil, err
		}

		serviceApiAuthPorts := []service.Port{
			{Name: "rest", Port: 80, TargetPort: "auth-rest"},
			{Name: "issuer", Port: 7070, TargetPort: "issuer-rest"},
			{Name: "grpc", Port: 8080, TargetPort: "auth-grpc"},
		}
		queryAA, destroyAA, err := service.AdaptFunc(k8sClient, "api-auth-v1", namespaceStr, labels, serviceApiAuthPorts, "", labels, false, "", "")
		if err != nil {
			return nil, nil, err
		}

		serviceApiMgmtPorts := []service.Port{
			{Name: "rest", Port: 80, TargetPort: "management-rest"},
			{Name: "grpc", Port: 8080, TargetPort: "management-grpc"},
		}
		querySAM, destroySAM, err := service.AdaptFunc(k8sClient, "api-management-v1", namespaceStr, labels, serviceApiMgmtPorts, "", labels, false, "", "")
		if err != nil {
			return nil, nil, err
		}

		serviceConsolePorts := []service.Port{
			{Name: "http", Port: 80, TargetPort: "console-http"},
		}
		querySC, destroySC, err := service.AdaptFunc(k8sClient, "console-v1", namespaceStr, labels, serviceConsolePorts, "", labels, false, "", "")
		if err != nil {
			return nil, nil, err
		}

		queriers = append(queriers, queryD, queryS, querySAA, queryAA, querySAM, querySC)
		destroyers = append(destroyers, destroyD, destroyS, destroySAA, destroyAA, destroySAM, destroySC)

		return func() (zitadel.EnsureFunc, error) {
				ensurers := make([]resources.EnsureFunc, 0)
				for _, querier := range queriers {
					ensurer, err := querier()
					if err != nil {
						return nil, err
					}
					ensurers = append(ensurers, ensurer)
				}

				return func() error {
					for _, ensurer := range ensurers {
						if err := ensurer(); err != nil {
							return err
						}
					}
					return nil
				}, nil
			}, func() error {
				for _, destroyer := range destroyers {
					if err := destroyer(); err != nil {
						return err
					}
				}
				return nil
			},
			nil
	}

}
