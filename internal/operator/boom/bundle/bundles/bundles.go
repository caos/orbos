package bundles

import (
	ambassadorinfo "github.com/caos/orbos/internal/operator/boom/application/applications/ambassador/info"
	argocdinfo "github.com/caos/orbos/internal/operator/boom/application/applications/argocd/info"
	grafanainfo "github.com/caos/orbos/internal/operator/boom/application/applications/grafana/info"
	kubestatemetricsinfo "github.com/caos/orbos/internal/operator/boom/application/applications/kubestatemetrics/info"
	loggingoperatorinfo "github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/info"
	lokiinfo "github.com/caos/orbos/internal/operator/boom/application/applications/loki/info"
	prometheusinfo "github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/info"
	prometheusnodeexporterinfo "github.com/caos/orbos/internal/operator/boom/application/applications/prometheusnodeexporter/info"
	prometheusoperatorinfo "github.com/caos/orbos/internal/operator/boom/application/applications/prometheusoperator/info"
	prometheussystemdexporterinfo "github.com/caos/orbos/internal/operator/boom/application/applications/prometheussystemdexporter/info"
	"github.com/caos/orbos/internal/operator/boom/name"
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

	return apps
}
