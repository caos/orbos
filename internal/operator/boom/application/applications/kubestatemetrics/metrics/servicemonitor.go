package metrics

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/kubestatemetrics/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/servicemonitor"
	"github.com/caos/orbos/internal/operator/boom/labels"
)

func GetServicemonitor(instanceName string) *servicemonitor.Config {
	appName := info.GetName()
	monitorlabels := labels.GetMonitorLabels(instanceName, appName)
	ls := labels.GetApplicationLabels(appName)

	relabelings := make([]*servicemonitor.ConfigRelabeling, 0)
	relabeling := &servicemonitor.ConfigRelabeling{
		Action: "labeldrop",
		Regex:  "(pod|service|endpoint|namespace)",
	}
	relabelings = append(relabelings, relabeling)

	endpoint := &servicemonitor.ConfigEndpoint{
		Port:        "http",
		Path:        "/metrics",
		HonorLabels: true,
		Relabelings: relabelings,
	}

	return &servicemonitor.Config{
		Name:                  "kube-state-metrics-servicemonitor",
		Endpoints:             []*servicemonitor.ConfigEndpoint{endpoint},
		MonitorMatchingLabels: monitorlabels,
		ServiceMatchingLabels: ls,
		JobName:               "kube-state-metrics",
	}
}
