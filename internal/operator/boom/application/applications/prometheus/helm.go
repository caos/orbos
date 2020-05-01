package prometheus

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta1"
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/config"
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/helm"
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/servicemonitor"
	"github.com/caos/orbos/internal/operator/boom/labels"
	"github.com/caos/orbos/internal/operator/boom/templator/helm/chart"
	"github.com/caos/orbos/internal/utils/kubectl"
	"github.com/caos/orbos/mntr"
)

func (p *Prometheus) SpecToHelmValues(monitor mntr.Monitor, toolsetCRDSpec *v1beta1.ToolsetSpec) interface{} {
	version, err := kubectl.NewVersion().GetKubeVersion(monitor)
	if err != nil {
		return nil
	}

	config := config.ScrapeMetricsCrdsConfig(info.GetInstanceName(), toolsetCRDSpec)

	values := helm.DefaultValues(p.GetImageTags())

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

	if config.MonitorLabels != nil {
		values.Prometheus.PrometheusSpec.ServiceMonitorSelector = &helm.MonitorSelector{
			MatchLabels: config.MonitorLabels,
		}
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

	ruleLabels := labels.GetRuleLabels(info.GetInstanceName())
	rules, _ := helm.GetDefaultRules(ruleLabels)

	values.Prometheus.PrometheusSpec.RuleSelector = &helm.RuleSelector{MatchLabels: ruleLabels}
	values.DefaultRules.Labels = ruleLabels
	values.KubeTargetVersionOverride = version
	values.AdditionalPrometheusRules = []*helm.AdditionalPrometheusRules{rules}

	values.FullnameOverride = info.GetInstanceName()

	return values
}

func (p *Prometheus) GetChartInfo() *chart.Chart {
	return helm.GetChartInfo()
}

func (p *Prometheus) GetImageTags() map[string]string {
	return helm.GetImageTags()
}
