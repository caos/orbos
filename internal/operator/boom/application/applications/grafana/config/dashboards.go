package config

import (
	"path/filepath"

	toolsetsv1beta2 "github.com/caos/orbos/internal/operator/boom/api/v1beta2"
)

var orgID = 0

func getGrafanaDashboards(dashboardsfolder string, toolsetCRDSpec *toolsetsv1beta2.ToolsetSpec) []*Provider {
	providers := make([]*Provider, 0)
	if toolsetCRDSpec.APIGateway != nil && toolsetCRDSpec.APIGateway.Deploy &&
		toolsetCRDSpec.MetricsPersisting != nil && (toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.Ambassador) {
		provider := &Provider{
			ConfigMaps: []string{
				"grafana-dashboard-ambassador-envoy-global",
				"grafana-dashboard-ambassador-envoy-ingress",
				"grafana-dashboard-ambassador-envoy-service",
			},
			Folder: filepath.Join(dashboardsfolder, "ambassador"),
		}
		providers = append(providers, provider)
	}

	if toolsetCRDSpec.Reconciling != nil && toolsetCRDSpec.Reconciling.Deploy &&
		toolsetCRDSpec.MetricsPersisting != nil && (toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.Argocd) {
		provider := &Provider{
			ConfigMaps: []string{
				"grafana-dashboard-argocd",
			},
			Folder: filepath.Join(dashboardsfolder, "argocd"),
		}
		providers = append(providers, provider)
	}

	nodeExporterDeployed := toolsetCRDSpec.NodeMetricsExporter != nil && toolsetCRDSpec.NodeMetricsExporter.Deploy &&
		toolsetCRDSpec.MetricsPersisting != nil && (toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.PrometheusNodeExporter)
	if nodeExporterDeployed {
		provider := &Provider{
			ConfigMaps: []string{
				"grafana-dashboard-node-cluster-rsrc-use",
				"grafana-dashboard-node-rsrc-use",
			},
			Folder: filepath.Join(dashboardsfolder, "prometheusnodeexporter"),
		}
		providers = append(providers, provider)
	}

	if toolsetCRDSpec.LogCollection != nil && toolsetCRDSpec.LogCollection.Deploy &&
		toolsetCRDSpec.MetricsPersisting != nil && (toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.LoggingOperator) {
		provider := &Provider{
			ConfigMaps: []string{
				"grafana-dashboard-logging-dashboard-rev3",
			},
			Folder: filepath.Join(dashboardsfolder, "loggingoperator"),
		}
		providers = append(providers, provider)
	}

	kubeStateMetricsDeployed := toolsetCRDSpec.KubeMetricsExporter != nil && toolsetCRDSpec.KubeMetricsExporter.Deploy &&
		toolsetCRDSpec.MetricsPersisting != nil && (toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.KubeStateMetrics)
	if kubeStateMetricsDeployed {
		provider := &Provider{
			ConfigMaps: []string{
				"grafana-persistentvolumesusage",
			},
			Folder: filepath.Join(dashboardsfolder, "persistentvolumesusage"),
		}
		providers = append(providers, provider)
	}

	if toolsetCRDSpec.MetricsPersisting != nil && (toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.Boom) {
		provider := &Provider{
			ConfigMaps: []string{
				"grafana-dashboard-boom",
			},
			Folder: filepath.Join(dashboardsfolder, "boom"),
		}
		providers = append(providers, provider)
	}

	if toolsetCRDSpec.MetricsPersisting != nil && (toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.Zitadel) {
		provider := &Provider{
			ConfigMaps: []string{
				"grafana-dashboard-zitadel-cockroachdb-replicas",
				"grafana-dashboard-zitadel-cockroachdb-runtime",
				"grafana-dashboard-zitadel-cockroachdb-sql",
				"grafana-dashboard-zitadel-cockroachdb-storage",
			},
			Folder: filepath.Join(dashboardsfolder, "zitadel"),
		}
		providers = append(providers, provider)
	}

	systemdExporterDeployed := toolsetCRDSpec.SystemdMetricsExporter != nil && toolsetCRDSpec.SystemdMetricsExporter.Deploy &&
		toolsetCRDSpec.MetricsPersisting != nil && (toolsetCRDSpec.MetricsPersisting.Metrics == nil || toolsetCRDSpec.MetricsPersisting.Metrics.PrometheusSystemdExporter)
	if systemdExporterDeployed && kubeStateMetricsDeployed && nodeExporterDeployed {
		provider := &Provider{
			ConfigMaps: []string{
				"grafana-dashboard-cluster-health",
				"grafana-dashboard-instance-health",
				"grafana-dashboard-probes-health",
			},
			Folder: filepath.Join(dashboardsfolder, "health"),
		}
		providers = append(providers, provider)
	}

	provider := &Provider{
		ConfigMaps: []string{
			"grafana-dashboard-kubelet",
		},
		Folder: filepath.Join(dashboardsfolder, "kubelet"),
	}
	providers = append(providers, provider)

	return providers
}
