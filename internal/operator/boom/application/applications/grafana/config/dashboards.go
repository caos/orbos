package config

import (
	"path/filepath"

	toolsetsv1beta1 "github.com/caos/orbos/internal/operator/boom/api/v1beta1"
)

var orgID = 0

func getGrafanaDashboards(dashboardsfolder string, toolsetCRDSpec *toolsetsv1beta1.ToolsetSpec) []*Provider {
	providers := make([]*Provider, 0)
	if toolsetCRDSpec.Ambassador != nil && toolsetCRDSpec.Ambassador.Deploy &&
		(toolsetCRDSpec.Prometheus.Metrics == nil || toolsetCRDSpec.Prometheus.Metrics.Ambassador) {
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

	if toolsetCRDSpec.Argocd != nil && toolsetCRDSpec.Argocd.Deploy &&
		(toolsetCRDSpec.Prometheus.Metrics == nil || toolsetCRDSpec.Prometheus.Metrics.Argocd) {
		provider := &Provider{
			ConfigMaps: []string{
				"grafana-dashboard-argocd",
			},
			Folder: filepath.Join(dashboardsfolder, "argocd"),
		}
		providers = append(providers, provider)
	}

	nodeExporterDeployed := toolsetCRDSpec.PrometheusNodeExporter != nil && toolsetCRDSpec.PrometheusNodeExporter.Deploy &&
		(toolsetCRDSpec.Prometheus.Metrics == nil || toolsetCRDSpec.Prometheus.Metrics.PrometheusNodeExporter)
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

	if toolsetCRDSpec.LoggingOperator != nil && toolsetCRDSpec.LoggingOperator.Deploy &&
		(toolsetCRDSpec.Prometheus.Metrics == nil || toolsetCRDSpec.Prometheus.Metrics.LoggingOperator) {
		provider := &Provider{
			ConfigMaps: []string{
				"grafana-dashboard-logging-dashboard-rev3",
			},
			Folder: filepath.Join(dashboardsfolder, "loggingoperator"),
		}
		providers = append(providers, provider)
	}

	kubeStateMetricsDeployed := toolsetCRDSpec.KubeStateMetrics != nil && toolsetCRDSpec.KubeStateMetrics.Deploy &&
		(toolsetCRDSpec.Prometheus.Metrics == nil || toolsetCRDSpec.Prometheus.Metrics.KubeStateMetrics)
	if kubeStateMetricsDeployed {
		provider := &Provider{
			ConfigMaps: []string{
				"grafana-persistentvolumesusage",
			},
			Folder: filepath.Join(dashboardsfolder, "persistentvolumesusage"),
		}
		providers = append(providers, provider)
	}

	if toolsetCRDSpec.Prometheus.Metrics == nil || toolsetCRDSpec.Prometheus.Metrics.Boom {
		provider := &Provider{
			ConfigMaps: []string{
				"grafana-dashboard-boom",
			},
			Folder: filepath.Join(dashboardsfolder, "boom"),
		}
		providers = append(providers, provider)
	}

	systemdExporterDeployed := toolsetCRDSpec.PrometheusSystemdExporter != nil && toolsetCRDSpec.PrometheusSystemdExporter.Deploy &&
		(toolsetCRDSpec.Prometheus.Metrics == nil || toolsetCRDSpec.Prometheus.Metrics.PrometheusSystemdExporter)
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
