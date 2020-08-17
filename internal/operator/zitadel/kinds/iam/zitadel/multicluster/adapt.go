package multicluster

import (
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/ambassador/host"
	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes/resources/ambassador/mapping"
	"github.com/caos/orbos/internal/operator/zitadel"
	coredb "github.com/caos/orbos/internal/operator/zitadel/kinds/databases/core"
	"github.com/caos/orbos/internal/operator/zitadel/kinds/networking/core"
	"github.com/caos/orbos/mntr"
)

func AdaptFunc(
	monitor mntr.Monitor,
	namespace string,
	labels map[string]string,
	dbSubdomain string,
) (
	zitadel.QueryFunc,
	zitadel.DestroyFunc,
	error,
) {
	internalMonitor := monitor.WithField("part", "database")

	dbMapping := "cockroachdb"
	hostName := dbMapping

	destroyHost, err := host.AdaptFuncToDestroy(namespace, hostName)
	if err != nil {
		return nil, nil, err
	}

	destroyMapping, err := mapping.AdaptFuncToDestroy(namespace, dbMapping)
	if err != nil {
		return nil, nil, err
	}

	destroyers := []zitadel.DestroyFunc{
		zitadel.ResourceDestroyToZitadelDestroy(destroyHost),
		zitadel.ResourceDestroyToZitadelDestroy(destroyMapping),
	}

	return func(k8sClient *kubernetes.Client, queried map[string]interface{}) (zitadel.EnsureFunc, error) {
			crd, err := k8sClient.CheckCRD("hosts.getambassador.io")
			if crd == nil || err != nil {
				return func(k8sClient *kubernetes.Client) error { return nil }, nil
			}

			crd, err = k8sClient.CheckCRD("mappings.getambassador.io")
			if crd == nil || err != nil {
				return func(k8sClient *kubernetes.Client) error { return nil }, nil
			}

			currentDB, err := coredb.ParseQueriedForDatabase(queried)
			if err != nil {
				return nil, err
			}

			currentNW, err := core.ParseQueriedForNetworking(queried)
			if err != nil {
				return nil, err
			}

			originCASecretName := currentNW.GetTlsCertName()

			dbDomain := dbSubdomain + "." + currentNW.GetDomain()
			selector := map[string]string{
				"hostname": dbDomain,
			}
			queryHost, err := host.AdaptFuncToEnsure(namespace, hostName, labels, dbDomain, "none", "", selector, originCASecretName)
			if err != nil {
				return nil, err
			}
			cors := &mapping.CORS{
				Origins:        "https://" + dbDomain,
				Methods:        "POST, GET, OPTIONS, DELETE, PUT",
				Headers:        "'*'",
				Credentials:    true,
				ExposedHeaders: "'*'",
				MaxAge:         "86400",
			}

			dbURL := currentDB.GetURL() + "." + namespace + ":" + currentDB.GetPort()
			queryMapping, err := mapping.AdaptFuncToEnsure(
				namespace,
				dbMapping,
				labels,
				true,
				dbDomain,
				"/",
				"",
				dbURL,
				"30000",
				"30000",
				cors,
			)
			if err != nil {
				return nil, err
			}

			queriers := []zitadel.QueryFunc{
				zitadel.EnsureFuncToQueryFunc(currentNW.GetReadyCertificate()),
				zitadel.ResourceQueryToZitadelQuery(queryHost),
				zitadel.ResourceQueryToZitadelQuery(queryMapping),
			}

			return zitadel.QueriersToEnsureFunc(internalMonitor, false, queriers, k8sClient, queried)
		},
		zitadel.DestroyersToDestroyFunc(internalMonitor, destroyers),
		nil
}
