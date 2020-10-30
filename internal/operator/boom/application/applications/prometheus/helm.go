package prometheus

import (
	"errors"
	"strconv"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/caos/orbos/internal/operator/boom/api/v1beta2"
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/config"
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/helm"
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/servicemonitor"
	"github.com/caos/orbos/internal/operator/boom/labels"
	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/internal/utils/clientgo"
	"github.com/caos/orbos/internal/utils/kubectl"
	"github.com/caos/orbos/mntr"
)

func (p *Prometheus) SpecToHelmValues(monitor mntr.Monitor, toolsetCRDSpec *v1beta2.ToolsetSpec) interface{} {
	version, err := kubectl.NewVersion().GetKubeVersion(monitor)
	if err != nil {
		monitor.Error(err)
		return nil
	}

	_, getSecretErr := clientgo.GetSecret("grafana-cloud", info.GetNamespace())
	ingestionSecretAbsent := k8serrors.IsNotFound(errors.Unwrap(getSecretErr))
	if getSecretErr != nil && !ingestionSecretAbsent {
		monitor.Info("Not sending telemetry data to MISSION as secret grafana-cloud is missing in namespace caos-system")
	}

	config := config.ScrapeMetricsCrdsConfig(info.GetInstanceName(), info.GetNamespace(), toolsetCRDSpec)

	values := helm.DefaultValues(p.GetImageTags())
	if config != nil {
		if config.StorageSpec != nil {
			storageSpec := &helm.StorageSpec{
				VolumeClaimTemplate: &helm.VolumeClaimTemplate{
					Spec: &helm.VolumeClaimTemplateSpec{
						StorageClassName: config.StorageSpec.StorageClass,
						AccessModes:      config.StorageSpec.AccessModes,
						Resources: &helm.Resources{
							Requests: &helm.Request{
								Storage: config.StorageSpec.Storage,
							},
						},
					},
				},
			}

			values.Prometheus.PrometheusSpec.StorageSpec = storageSpec
		}

		if config.ServiceMonitors != nil {
			additionalServiceMonitors := make([]*servicemonitor.Values, 0)
			for _, specServiceMonitor := range config.ServiceMonitors {
				valuesServiceMonitor := servicemonitor.SpecToValues(specServiceMonitor)
				additionalServiceMonitors = append(additionalServiceMonitors, valuesServiceMonitor)
			}

			values.Prometheus.AdditionalServiceMonitors = additionalServiceMonitors
		}

		if config.AdditionalScrapeConfigs != nil {
			values.Prometheus.PrometheusSpec.AdditionalScrapeConfigs = config.AdditionalScrapeConfigs
		}
	}

	values.Prometheus.PrometheusSpec.ExternalLabels = make(map[string]string)
	if toolsetCRDSpec.MetricsPersisting.ExternalLabels != nil {
		for k, v := range toolsetCRDSpec.MetricsPersisting.ExternalLabels {
			if k == "orb" {
				monitor.Info("Label-key \"orb\" is already used internally and will be ignored")
			} else {
				values.Prometheus.PrometheusSpec.ExternalLabels[k] = v
			}
		}
	}

	if getSecretErr == nil && !ingestionSecretAbsent {
		values.Prometheus.PrometheusSpec.ExternalLabels["orb"] = p.orb
		values.Prometheus.PrometheusSpec.RemoteWrite = append(values.Prometheus.PrometheusSpec.RemoteWrite, &helm.RemoteWrite{
			URL: "https://prometheus-us-central1.grafana.net/api/prom/push",
			BasicAuth: &helm.BasicAuth{
				Username: &helm.SecretKeySelector{
					Name: "grafana-cloud",
					Key:  "username",
				},
				Password: &helm.SecretKeySelector{
					Name: "grafana-cloud",
					Key:  "password",
				},
			},
			WriteRelabelConfigs: []*helm.ValuesRelabelConfig{{
				Action: "keep",
				SourceLabels: []string{
					"__name__",
					"job",
				},
				Regex: "caos_.+;.*|up;caos_remote_.+",
			}},
		})
	}

	if toolsetCRDSpec.MetricsPersisting != nil && toolsetCRDSpec.MetricsPersisting.RemoteWrite != nil {
		writeRelabelConfigs := make([]*helm.ValuesRelabelConfig, 0)
		if toolsetCRDSpec.MetricsPersisting.RemoteWrite.RelabelConfigs != nil && len(toolsetCRDSpec.MetricsPersisting.RemoteWrite.RelabelConfigs) > 0 {
			for _, relabelConfig := range toolsetCRDSpec.MetricsPersisting.RemoteWrite.RelabelConfigs {
				mod := 0
				if relabelConfig.Modulus != "" {
					internalMod, err := strconv.Atoi(relabelConfig.Modulus)
					if err != nil {
						return err
					}
					mod = internalMod
				}

				writeRelabelConfigs = append(writeRelabelConfigs, &helm.ValuesRelabelConfig{
					Action:       relabelConfig.Action,
					SourceLabels: relabelConfig.SourceLabels,
					Separator:    relabelConfig.Separator,
					TargetLabel:  relabelConfig.TargetLabel,
					Regex:        relabelConfig.Regex,
					Modulus:      uint64(mod),
					Replacement:  relabelConfig.Replacement,
				})
			}
		}

		values.Prometheus.PrometheusSpec.RemoteWrite = append(values.Prometheus.PrometheusSpec.RemoteWrite, &helm.RemoteWrite{
			URL: toolsetCRDSpec.MetricsPersisting.RemoteWrite.URL,
			BasicAuth: &helm.BasicAuth{
				Username: &helm.SecretKeySelector{
					Name: toolsetCRDSpec.MetricsPersisting.RemoteWrite.BasicAuth.Username.Name,
					Key:  toolsetCRDSpec.MetricsPersisting.RemoteWrite.BasicAuth.Username.Key,
				},
				Password: &helm.SecretKeySelector{
					Name: toolsetCRDSpec.MetricsPersisting.RemoteWrite.BasicAuth.Password.Name,
					Key:  toolsetCRDSpec.MetricsPersisting.RemoteWrite.BasicAuth.Password.Key,
				},
			},
			WriteRelabelConfigs: writeRelabelConfigs,
		})
	}

	if toolsetCRDSpec.MetricsPersisting.Tolerations != nil {
		for _, tol := range toolsetCRDSpec.MetricsPersisting.Tolerations {
			values.Prometheus.PrometheusSpec.Tolerations = append(values.Prometheus.PrometheusSpec.Tolerations, tol)
		}
	}

	promSelectorLabels := labels.GetPromSelector(info.GetInstanceName())
	promSelector := &helm.Selector{MatchLabels: promSelectorLabels}
	resourceLabels := labels.GetRuleLabels(info.GetInstanceName(), info.GetName())

	values.Prometheus.PrometheusSpec.RuleSelector = promSelector
	values.Prometheus.PrometheusSpec.PodMonitorSelector = promSelector
	values.Prometheus.PrometheusSpec.ServiceMonitorSelector = promSelector

	rules, err := helm.GetDefaultRules(resourceLabels)
	if err != nil {
		panic(err)
	}
	values.DefaultRules.Labels = resourceLabels
	values.KubeTargetVersionOverride = version
	values.AdditionalPrometheusRules = []*helm.AdditionalPrometheusRules{rules}

	values.FullnameOverride = info.GetInstanceName()

	if toolsetCRDSpec.MetricsPersisting.NodeSelector != nil {
		for k, v := range toolsetCRDSpec.MetricsPersisting.NodeSelector {
			values.Prometheus.PrometheusSpec.NodeSelector[k] = v
		}
	}

	if toolsetCRDSpec.MetricsPersisting.Resources != nil {
		values.Prometheus.PrometheusSpec.Resources = toolsetCRDSpec.MetricsPersisting.Resources
	}

	return values
}

func (p *Prometheus) GetChartInfo() *chart.Chart {
	return helm.GetChartInfo()
}

func (p *Prometheus) GetImageTags() map[string]string {
	return helm.GetImageTags()
}
