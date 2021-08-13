package config

import (
	toolsetslatest "github.com/caos/orbos/internal/operator/boom/api/latest"
	ambassadormetrics "github.com/caos/orbos/internal/operator/boom/application/applications/apigateway/metrics"
	"github.com/caos/orbos/internal/operator/boom/application/applications/apiserver"
	"github.com/caos/orbos/internal/operator/boom/application/applications/boom"
	"github.com/caos/orbos/internal/operator/boom/application/applications/database"
	kubestatemetrics "github.com/caos/orbos/internal/operator/boom/application/applications/kubemetricsexporter/metrics"
	lometrics "github.com/caos/orbos/internal/operator/boom/application/applications/logcollection/metrics"
	lokimetrics "github.com/caos/orbos/internal/operator/boom/application/applications/logspersisting/metrics"
	pometrics "github.com/caos/orbos/internal/operator/boom/application/applications/metriccollection/metrics"
	"github.com/caos/orbos/internal/operator/boom/application/applications/metricspersisting/metrics"
	"github.com/caos/orbos/internal/operator/boom/application/applications/metricspersisting/servicemonitor"
	"github.com/caos/orbos/internal/operator/boom/application/applications/networking"
	pnemetrics "github.com/caos/orbos/internal/operator/boom/application/applications/nodemetricsexporter/metrics"
	"github.com/caos/orbos/internal/operator/boom/application/applications/orbiter"
	argocdmetrics "github.com/caos/orbos/internal/operator/boom/application/applications/reconciling/metrics"
	psemetrics "github.com/caos/orbos/internal/operator/boom/application/applications/systemdmetricsexporter/metrics"
	"github.com/caos/orbos/internal/operator/boom/application/applications/zitadel"
	"github.com/caos/orbos/internal/operator/boom/labels"
)

func ScrapeMetricsCrdsConfig(
	instanceName string,
	namespace string,
	toolsetCRDSpec *toolsetslatest.ToolsetSpec,
	withGrafanaCloud bool,
) *Config {
	if toolsetCRDSpec.MetricsPersisting == nil || !toolsetCRDSpec.MetricsPersisting.Deploy {
		return nil
	}
	servicemonitors := make([]*servicemonitor.Config, 0)

	if toolsetCRDSpec.APIGateway != nil && toolsetCRDSpec.APIGateway.Deploy &&
		(toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.Ambassador) {
		servicemonitors = append(servicemonitors, ambassadormetrics.GetServicemonitor(instanceName))
		if withGrafanaCloud {
			servicemonitors = append(servicemonitors, ambassadormetrics.GetCloudServicemonitor(instanceName))
		}
	}

	if toolsetCRDSpec.MetricCollection != nil && toolsetCRDSpec.MetricCollection.Deploy &&
		(toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.PrometheusOperator) {
		servicemonitors = append(servicemonitors, pometrics.GetServicemonitor(instanceName))
	}

	if toolsetCRDSpec.NodeMetricsExporter != nil && toolsetCRDSpec.NodeMetricsExporter.Deploy &&
		(toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.PrometheusNodeExporter) {
		servicemonitors = append(servicemonitors, pnemetrics.GetServicemonitors(instanceName)...)
	}

	if toolsetCRDSpec.SystemdMetricsExporter != nil && toolsetCRDSpec.SystemdMetricsExporter.Deploy &&
		(toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.PrometheusSystemdExporter) {
		servicemonitors = append(servicemonitors, psemetrics.GetServicemonitor(instanceName))
	}

	if toolsetCRDSpec.KubeMetricsExporter != nil && toolsetCRDSpec.KubeMetricsExporter.Deploy &&
		(toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.KubeStateMetrics) {
		servicemonitors = append(servicemonitors, kubestatemetrics.GetServicemonitors(instanceName)...)
	}

	if toolsetCRDSpec.Reconciling != nil && toolsetCRDSpec.Reconciling.Deploy &&
		(toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.Argocd) {
		servicemonitors = append(servicemonitors, argocdmetrics.GetServicemonitors(instanceName)...)
	}

	if toolsetCRDSpec.LogCollection != nil && toolsetCRDSpec.LogCollection.Deploy &&
		(toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.LoggingOperator) {
		servicemonitors = append(servicemonitors, lometrics.GetServicemonitors(instanceName)...)
	}

	if toolsetCRDSpec.LogsPersisting != nil && toolsetCRDSpec.LogsPersisting.Deploy &&
		(toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.Loki) {
		servicemonitors = append(servicemonitors, lokimetrics.GetServicemonitor(instanceName))
	}

	if toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.Prometheus {
		servicemonitors = append(servicemonitors, metrics.GetServicemonitor(instanceName))
	}

	if toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.APIServer {
		servicemonitors = append(servicemonitors, apiserver.GetServicemonitor(instanceName))
	}

	if toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.Boom {
		servicemonitors = append(servicemonitors, boom.GetServicemonitor(instanceName))
	}

	if toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.Orbiter {
		servicemonitors = append(servicemonitors, orbiter.GetServicemonitor(instanceName))
	}

	if toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.Zitadel {
		for _, sm := range zitadel.GetServicemonitors(instanceName) {
			servicemonitors = append(servicemonitors, sm)
		}
		if withGrafanaCloud {
			servicemonitors = append(servicemonitors, zitadel.GetCloudZitadelServiceMonitor(instanceName))
		}
	}

	if toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.Database {
		for _, sm := range database.GetServicemonitors(instanceName) {
			servicemonitors = append(servicemonitors, sm)
		}
		if withGrafanaCloud {
			servicemonitors = append(servicemonitors, database.GetCloudDatabaseServiceMonitor(instanceName))
		}
	}

	if toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.Networking {
		servicemonitors = append(servicemonitors, networking.GetServicemonitor(instanceName))
	}

	prom := &Config{
		Prefix:                  "",
		Namespace:               namespace,
		MonitorLabels:           labels.GetMonitorSelectorLabels(instanceName),
		RuleLabels:              labels.GetRuleSelectorLabels(instanceName),
		ServiceMonitors:         servicemonitors,
		AdditionalScrapeConfigs: getScrapeConfigs(),
	}

	if toolsetCRDSpec.MetricsPersisting.Storage != nil {
		prom.StorageSpec = &StorageSpec{
			StorageClass: toolsetCRDSpec.MetricsPersisting.Storage.StorageClass,
			Storage:      toolsetCRDSpec.MetricsPersisting.Storage.Size,
		}

		if toolsetCRDSpec.MetricsPersisting.Storage.AccessModes != nil {
			prom.StorageSpec.AccessModes = toolsetCRDSpec.MetricsPersisting.Storage.AccessModes
		}
	}

	return prom
}
