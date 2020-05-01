package metrics

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/servicemonitor"
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheusnodeexporter/info"
	"github.com/caos/orbos/internal/operator/boom/labels"
)

func GetServicemonitor(instanceName string) *servicemonitor.Config {
	appName := info.GetName()
	monitorlabels := labels.GetMonitorLabels(instanceName, appName)
	ls := labels.GetApplicationLabels(appName)

	relabelings := make([]*servicemonitor.ConfigRelabeling, 0)
	relabeling := &servicemonitor.ConfigRelabeling{
		Action:       "replace",
		Regex:        "(.*)",
		Replacement:  "$1",
		SourceLabels: []string{"__meta_kubernetes_pod_node_name"},
		TargetLabel:  "instance",
	}
	relabelings = append(relabelings, relabeling)

	endpoint := &servicemonitor.ConfigEndpoint{
		Port:        "metrics",
		Path:        "/metrics",
		Relabelings: relabelings,
	}

	return &servicemonitor.Config{
		Name:                  "prometheus-node-exporter-servicemonitor",
		Endpoints:             []*servicemonitor.ConfigEndpoint{endpoint},
		MonitorMatchingLabels: monitorlabels,
		ServiceMatchingLabels: ls,
		JobName:               appName.String(),
	}
}
