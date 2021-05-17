package application

import (
	"github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/internal/operator/boom/application/applications/ambassador"
	ambassadorinfo "github.com/caos/orbos/internal/operator/boom/application/applications/ambassador/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/argocd"
	argocdinfo "github.com/caos/orbos/internal/operator/boom/application/applications/argocd/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/grafana"
	grafanainfo "github.com/caos/orbos/internal/operator/boom/application/applications/grafana/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/kubestatemetrics"
	kubestatemetricsinfo "github.com/caos/orbos/internal/operator/boom/application/applications/kubestatemetrics/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator"
	loggingoperatorinfo "github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/loki"
	lokiinfo "github.com/caos/orbos/internal/operator/boom/application/applications/loki/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/metricsserver"
	metricsserverinfo "github.com/caos/orbos/internal/operator/boom/application/applications/metricsserver/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/miniooperator"
	miniooperatorinfo "github.com/caos/orbos/internal/operator/boom/application/applications/miniooperator/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheus"
	prometheusinfo "github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheusnodeexporter"
	prometheusnodeexporterinfo "github.com/caos/orbos/internal/operator/boom/application/applications/prometheusnodeexporter/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheusoperator"
	prometheusoperatorinfo "github.com/caos/orbos/internal/operator/boom/application/applications/prometheusoperator/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheussystemdexporter"
	prometheussystemdexporterinfo "github.com/caos/orbos/internal/operator/boom/application/applications/prometheussystemdexporter/info"
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/mntr"
)

type Application interface {
	Deploy(*latest.ToolsetSpec) bool
	GetName() name.Application
}

type HelmApplication interface {
	Application
	GetNamespace() string
	GetChartInfo() *chart.Chart
	GetImageTags() map[string]string
	SpecToHelmValues(mntr.Monitor, *latest.ToolsetSpec) interface{}
}

type YAMLApplication interface {
	Application
	GetYaml(mntr.Monitor, *latest.ToolsetSpec) interface{}
}

func New(monitor mntr.Monitor, appName name.Application, orb string) Application {
	switch appName {
	case ambassadorinfo.GetName():
		return ambassador.New(monitor)
	case argocdinfo.GetName():
		return argocd.New(monitor)
	case grafanainfo.GetName():
		return grafana.New(monitor)
	case kubestatemetricsinfo.GetName():
		return kubestatemetrics.New(monitor)
	case prometheusoperatorinfo.GetName():
		return prometheusoperator.New(monitor)
	case loggingoperatorinfo.GetName():
		return loggingoperator.New(monitor)
	case prometheusnodeexporterinfo.GetName():
		return prometheusnodeexporter.New(monitor)
	case prometheussystemdexporterinfo.GetName():
		return prometheussystemdexporter.New()
	case prometheusinfo.GetName():
		return prometheus.New(monitor, orb)
	case lokiinfo.GetName():
		return loki.New(monitor)
	case metricsserverinfo.GetName():
		return metricsserver.New(monitor)
	case miniooperatorinfo.GetName():
		return miniooperator.New(monitor)
	}

	return nil
}

func GetOrderNumber(appName name.Application) int {
	switch appName {
	case prometheusinfo.GetName():
		return prometheusinfo.GetOrderNumber()
	case lokiinfo.GetName():
		return lokiinfo.GetOrderNumber()
	}

	return 1
}
