package bundles

import (
	ambassadorinfo "github.com/caos/orbos/v5/internal/operator/boom/application/applications/apigateway/info"
	kubestatemetricsinfo "github.com/caos/orbos/v5/internal/operator/boom/application/applications/kubemetricsexporter/info"
	loggingoperatorinfo "github.com/caos/orbos/v5/internal/operator/boom/application/applications/logcollection/info"
	lokiinfo "github.com/caos/orbos/v5/internal/operator/boom/application/applications/logspersisting/info"
	prometheusoperatorinfo "github.com/caos/orbos/v5/internal/operator/boom/application/applications/metriccollection/info"
	prometheusinfo "github.com/caos/orbos/v5/internal/operator/boom/application/applications/metricspersisting/info"
	metricsserverinfo "github.com/caos/orbos/v5/internal/operator/boom/application/applications/metricsserver/info"
	grafanainfo "github.com/caos/orbos/v5/internal/operator/boom/application/applications/monitoring/info"
	prometheusnodeexporterinfo "github.com/caos/orbos/v5/internal/operator/boom/application/applications/nodemetricsexporter/info"
	argocdinfo "github.com/caos/orbos/v5/internal/operator/boom/application/applications/reconciling/info"
	prometheussystemdexporterinfo "github.com/caos/orbos/v5/internal/operator/boom/application/applications/systemdmetricsexporter/info"
	"github.com/caos/orbos/v5/internal/operator/boom/name"
)

const (
	Caos  name.Bundle = "caos"
	Empty name.Bundle = "empty"
)

func GetAll() []name.Application {
	apps := make([]name.Application, 0)
	apps = append(apps, GetCaos()...)
	return apps
}

func Get(bundle name.Bundle) []name.Application {
	switch bundle {
	case Caos:
		return GetCaos()
	case Empty:
		return make([]name.Application, 0)
	}

	return nil
}

func GetCaos() []name.Application {

	apps := make([]name.Application, 0)
	apps = append(apps, ambassadorinfo.GetName())
	apps = append(apps, argocdinfo.GetName())
	apps = append(apps, prometheusoperatorinfo.GetName())
	apps = append(apps, kubestatemetricsinfo.GetName())
	apps = append(apps, prometheusnodeexporterinfo.GetName())
	apps = append(apps, prometheussystemdexporterinfo.GetName())
	apps = append(apps, grafanainfo.GetName())
	apps = append(apps, prometheusinfo.GetName())
	apps = append(apps, loggingoperatorinfo.GetName())
	apps = append(apps, lokiinfo.GetName())
	apps = append(apps, metricsserverinfo.GetName())

	return apps
}
