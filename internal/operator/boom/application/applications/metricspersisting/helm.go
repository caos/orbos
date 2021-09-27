package metricspersisting

import (
	"errors"
	"strconv"

	"github.com/caos/orbos/v5/internal/utils/helper"

	k8serrors "k8s.io/apimachinery/pkg/api/errors"

	"github.com/caos/orbos/v5/internal/operator/boom/api/latest"
	"github.com/caos/orbos/v5/internal/operator/boom/application/applications/metricspersisting/config"
	"github.com/caos/orbos/v5/internal/operator/boom/application/applications/metricspersisting/helm"
	"github.com/caos/orbos/v5/internal/operator/boom/application/applications/metricspersisting/info"
	"github.com/caos/orbos/v5/internal/operator/boom/application/applications/metricspersisting/servicemonitor"
	"github.com/caos/orbos/v5/internal/operator/boom/labels"
	"github.com/caos/orbos/v5/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/v5/internal/utils/clientgo"
	"github.com/caos/orbos/v5/internal/utils/kubectl"
	"github.com/caos/orbos/v5/mntr"
)

func (p *Prometheus) SpecToHelmValues(monitor mntr.Monitor, toolsetCRDSpec *latest.ToolsetSpec) interface{} {
	version, err := kubectl.NewVersion().GetKubeVersion(monitor)
	if err != nil {
		monitor.Error(err)
		return nil
	}

	_, getSecretErr := clientgo.GetSecret("grafana-cloud", info.GetNamespace())
	telemetrySecretAbsent := k8serrors.IsNotFound(errors.Unwrap(getSecretErr))
	if getSecretErr != nil && !telemetrySecretAbsent {
		monitor.Info("Not sending telemetry data to MISSION as secret grafana-cloud is missing in namespace caos-system")
	}

	configResult := config.ScrapeMetricsCrdsConfig(info.GetInstanceName(), info.GetNamespace(), toolsetCRDSpec)
	if configResult == nil {
		return nil
	}

	imageTags := p.GetImageTags()
	image := "quay.io/prometheus/prometheus"

	if toolsetCRDSpec != nil && toolsetCRDSpec.MetricsPersisting != nil {
		helper.OverwriteExistingValues(imageTags, map[string]string{
			image: toolsetCRDSpec.MetricsPersisting.OverwriteVersion,
		})
		helper.OverwriteExistingKey(imageTags, &image, toolsetCRDSpec.MetricsPersisting.OverwriteImage)
	}

	values := helm.DefaultValues(imageTags, image)
	if configResult.StorageSpec != nil {
		storageSpec := &helm.StorageSpec{
			VolumeClaimTemplate: &helm.VolumeClaimTemplate{
				Spec: &helm.VolumeClaimTemplateSpec{
					StorageClassName: configResult.StorageSpec.StorageClass,
					AccessModes:      configResult.StorageSpec.AccessModes,
					Resources: &helm.Resources{
						Requests: &helm.Request{
							Storage: configResult.StorageSpec.Storage,
						},
					},
				},
			},
		}

		values.Prometheus.PrometheusSpec.StorageSpec = storageSpec
	}

	if configResult.ServiceMonitors != nil {
		additionalServiceMonitors := make([]*servicemonitor.Values, 0)
		for _, specServiceMonitor := range configResult.ServiceMonitors {
			valuesServiceMonitor := servicemonitor.SpecToValues(specServiceMonitor)
			additionalServiceMonitors = append(additionalServiceMonitors, valuesServiceMonitor)
		}

		values.Prometheus.AdditionalServiceMonitors = additionalServiceMonitors
	}

	if configResult.AdditionalScrapeConfigs != nil {
		values.Prometheus.PrometheusSpec.AdditionalScrapeConfigs = configResult.AdditionalScrapeConfigs
	}

	spec := toolsetCRDSpec.MetricsPersisting
	if spec == nil {
		return values
	}

	values.Prometheus.PrometheusSpec.ExternalLabels = make(map[string]string)
	if spec.ExternalLabels != nil {
		for k, v := range spec.ExternalLabels {
			if k == "orb" {
				monitor.Info("Label-key \"orb\" is already used internally and will be ignored")
			} else {
				values.Prometheus.PrometheusSpec.ExternalLabels[k] = v
			}
		}
	}

	if getSecretErr == nil && !telemetrySecretAbsent {
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

	if spec.RemoteWrite != nil {
		writeRelabelConfigs := make([]*helm.ValuesRelabelConfig, 0)
		if spec.RemoteWrite.RelabelConfigs != nil && len(spec.RemoteWrite.RelabelConfigs) > 0 {
			for _, relabelConfig := range spec.RemoteWrite.RelabelConfigs {
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
			URL: spec.RemoteWrite.URL,
			BasicAuth: &helm.BasicAuth{
				Username: &helm.SecretKeySelector{
					Name: spec.RemoteWrite.BasicAuth.Username.Name,
					Key:  spec.RemoteWrite.BasicAuth.Username.Key,
				},
				Password: &helm.SecretKeySelector{
					Name: spec.RemoteWrite.BasicAuth.Password.Name,
					Key:  spec.RemoteWrite.BasicAuth.Password.Key,
				},
			},
			WriteRelabelConfigs: writeRelabelConfigs,
		})
	}

	if spec.Tolerations != nil {
		for _, tol := range spec.Tolerations {
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

	if spec.NodeSelector != nil {
		for k, v := range spec.NodeSelector {
			values.Prometheus.PrometheusSpec.NodeSelector[k] = v
		}
	}

	if spec.Resources != nil {
		values.Prometheus.PrometheusSpec.Resources = spec.Resources
	}

	return values
}

func (p *Prometheus) GetChartInfo() *chart.Chart {
	return helm.GetChartInfo()
}

func (p *Prometheus) GetImageTags() map[string]string {
	return helm.GetImageTags()
}
