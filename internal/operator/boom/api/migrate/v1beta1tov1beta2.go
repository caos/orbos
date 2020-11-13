package migrate

import (
	"github.com/caos/orbos/internal/operator/boom/api/latest"
	"github.com/caos/orbos/internal/operator/boom/api/migrate/argocd"
	"github.com/caos/orbos/internal/operator/boom/api/migrate/grafana"
	"github.com/caos/orbos/internal/operator/boom/api/migrate/storage"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	"github.com/caos/orbos/internal/secret"
)

func V1beta1Tov1beta2(oldToolset *v1beta1.Toolset) (*v1beta2.Toolset, map[string]*secret.Secret) {
	newToolset := &v1beta2.Toolset{
		APIVersion: "boom.caos.ch/v1beta2",
		Metadata: &latest.Metadata{
			Name:      oldToolset.Metadata.Name,
			Namespace: oldToolset.Metadata.Namespace,
		},
		Kind: "Toolset",
	}
	if oldToolset.Spec != nil {
		oldSpec := oldToolset.Spec
		newSpec := &v1beta2.ToolsetSpec{
			CurrentStateFolder: oldSpec.CurrentStateFolder,
			ForceApply:         oldSpec.ForceApply,
		}
		if oldSpec.BoomVersion != "" {
			newSpec.Boom = &latest.Boom{
				Version: oldSpec.BoomVersion,
			}
		}

		if oldSpec.PreApply != nil {
			newSpec.PreApply = &latest.Apply{
				Deploy: oldSpec.PreApply.Deploy,
				Folder: oldSpec.PreApply.Folder,
			}
		}

		if oldSpec.PostApply != nil {
			newSpec.PostApply = &latest.Apply{
				Deploy: oldSpec.PostApply.Deploy,
				Folder: oldSpec.PostApply.Folder,
			}
		}

		if oldSpec.PrometheusOperator != nil {
			newSpec.MetricCollection = &latest.MetricCollection{
				Deploy:       oldSpec.PrometheusOperator.Deploy,
				NodeSelector: oldSpec.PrometheusOperator.NodeSelector,
			}
		}

		if oldSpec.LoggingOperator != nil {
			newSpec.LogCollection = &v1beta2.LogCollection{
				Deploy:       oldSpec.LoggingOperator.Deploy,
				NodeSelector: oldSpec.LoggingOperator.NodeSelector,
			}
			if oldSpec.LoggingOperator.FluentdPVC != nil {
				newSpec.LogCollection.FluentdPVC = storage.V1beta1Tov1beta2(oldSpec.LoggingOperator.FluentdPVC)
			}
		}

		if oldSpec.PrometheusNodeExporter != nil {
			newSpec.NodeMetricsExporter = &latest.NodeMetricsExporter{Deploy: oldSpec.PrometheusNodeExporter.Deploy}
		}

		if oldSpec.PrometheusSystemdExporter != nil {
			newSpec.SystemdMetricsExporter = &latest.SystemdMetricsExporter{Deploy: oldSpec.PrometheusSystemdExporter.Deploy}
		}

		if oldSpec.Ambassador != nil {
			newSpec.APIGateway = &latest.APIGateway{
				Deploy:            oldSpec.Ambassador.Deploy,
				ReplicaCount:      oldSpec.Ambassador.ReplicaCount,
				ActivateDevPortal: oldSpec.Ambassador.ActivateDevPortal,
				NodeSelector:      oldSpec.Ambassador.NodeSelector,
			}
			if oldSpec.Ambassador.Service != nil {
				newSpec.APIGateway.Service = &latest.AmbassadorService{
					Type:           oldSpec.Ambassador.Service.Type,
					LoadBalancerIP: oldSpec.Ambassador.Service.LoadBalancerIP,
				}
				if oldSpec.Ambassador.Service.Ports != nil {
					ports := make([]*latest.Port, 0)
					for _, v := range oldSpec.Ambassador.Service.Ports {
						ports = append(ports, &latest.Port{
							Name:       v.Name,
							Port:       v.Port,
							TargetPort: v.TargetPort,
							NodePort:   v.NodePort,
						})
					}
					newSpec.APIGateway.Service.Ports = ports
				}
			}
		}

		if oldSpec.Grafana != nil {
			newSpec.Monitoring = grafana.V1beta1Tov1beta2(oldSpec.Grafana)
		}

		if oldSpec.Argocd != nil {
			newSpec.Reconciling = argocd.V1beta1Tov1beta2(oldSpec.Argocd)
		}
		if oldSpec.KubeStateMetrics != nil {
			newSpec.KubeMetricsExporter = &latest.KubeMetricsExporter{
				Deploy:       oldSpec.KubeStateMetrics.Deploy,
				ReplicaCount: oldSpec.KubeStateMetrics.ReplicaCount,
				NodeSelector: oldSpec.KubeStateMetrics.NodeSelector,
			}
		}
		if oldSpec.Prometheus != nil {
			newSpec.MetricsPersisting = &latest.MetricsPersisting{
				Deploy:       oldSpec.Prometheus.Deploy,
				NodeSelector: oldSpec.Prometheus.NodeSelector,
			}
			if oldSpec.Prometheus.Storage != nil {
				newSpec.MetricsPersisting.Storage = storage.V1beta1Tov1beta2(oldSpec.Prometheus.Storage)
			}
			if oldSpec.Prometheus.Metrics != nil {
				newSpec.MetricsPersisting.Metrics = &latest.Metrics{
					Ambassador:                oldSpec.Prometheus.Metrics.Ambassador,
					Argocd:                    oldSpec.Prometheus.Metrics.Argocd,
					KubeStateMetrics:          oldSpec.Prometheus.Metrics.KubeStateMetrics,
					PrometheusNodeExporter:    oldSpec.Prometheus.Metrics.PrometheusNodeExporter,
					PrometheusSystemdExporter: oldSpec.Prometheus.Metrics.PrometheusSystemdExporter,
					APIServer:                 oldSpec.Prometheus.Metrics.APIServer,
					PrometheusOperator:        oldSpec.Prometheus.Metrics.PrometheusOperator,
					LoggingOperator:           oldSpec.Prometheus.Metrics.LoggingOperator,
					Loki:                      oldSpec.Prometheus.Metrics.Loki,
					Boom:                      oldSpec.Prometheus.Metrics.Boom,
					Orbiter:                   oldSpec.Prometheus.Metrics.Orbiter,
				}
			}
			if oldSpec.Prometheus.RemoteWrite != nil {
				newSpec.MetricsPersisting.RemoteWrite = &latest.RemoteWrite{
					URL: oldSpec.Prometheus.RemoteWrite.URL,
				}
				if oldSpec.Prometheus.RemoteWrite.BasicAuth != nil {
					newSpec.MetricsPersisting.RemoteWrite.BasicAuth = &latest.BasicAuth{}
					if oldSpec.Prometheus.RemoteWrite.BasicAuth.Username != nil {
						newSpec.MetricsPersisting.RemoteWrite.BasicAuth.Username = &latest.SecretKeySelector{
							Name: oldSpec.Prometheus.RemoteWrite.BasicAuth.Username.Name,
							Key:  oldSpec.Prometheus.RemoteWrite.BasicAuth.Username.Key,
						}
					}
					if oldSpec.Prometheus.RemoteWrite.BasicAuth.Password != nil {
						newSpec.MetricsPersisting.RemoteWrite.BasicAuth.Password = &latest.SecretKeySelector{
							Name: oldSpec.Prometheus.RemoteWrite.BasicAuth.Password.Name,
							Key:  oldSpec.Prometheus.RemoteWrite.BasicAuth.Password.Key,
						}
					}
				}
			}
		}
		if oldSpec.Loki != nil {
			newSpec.LogsPersisting = &latest.LogsPersisting{
				Deploy:        oldSpec.Loki.Deploy,
				ClusterOutput: oldSpec.Loki.ClusterOutput,
				NodeSelector:  oldSpec.Loki.NodeSelector,
			}
			if oldSpec.Loki != nil {
				newSpec.LogsPersisting.Storage = storage.V1beta1Tov1beta2(oldSpec.Loki.Storage)
			}
			if oldSpec.Loki.Logs != nil {
				newSpec.LogsPersisting.Logs = &latest.Logs{
					Ambassador:                oldSpec.Loki.Logs.Ambassador,
					Grafana:                   oldSpec.Loki.Logs.Grafana,
					Argocd:                    oldSpec.Loki.Logs.Argocd,
					KubeStateMetrics:          oldSpec.Loki.Logs.KubeStateMetrics,
					PrometheusNodeExporter:    oldSpec.Loki.Logs.PrometheusNodeExporter,
					PrometheusOperator:        oldSpec.Loki.Logs.PrometheusOperator,
					LoggingOperator:           oldSpec.Loki.Logs.LoggingOperator,
					Loki:                      oldSpec.Loki.Logs.Loki,
					Prometheus:                oldSpec.Loki.Logs.Prometheus,
					MetricsServer:             oldSpec.Loki.Logs.MetricsServer,
					PrometheusSystemdExporter: oldSpec.Loki.Logs.PrometheusSystemdExporter,
				}
			}
		}
		if oldSpec.MetricsServer != nil {
			newSpec.MetricsServer = &latest.MetricsServer{
				Deploy: oldSpec.MetricsServer.Deploy,
			}
		}
		newToolset.Spec = newSpec
	}

	return newToolset, v1beta2.GetSecretsMap(newToolset)
}
