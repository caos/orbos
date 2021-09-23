package metrics

import (
	"fmt"

	"github.com/caos/orbos/internal/operator/boom/application/applications/metricspersisting/servicemonitor"
	"github.com/caos/orbos/internal/operator/boom/application/applications/systemdmetricsexporter/info"
	"github.com/caos/orbos/internal/operator/boom/labels"
)

func GetServicemonitor(instanceName string) *servicemonitor.Config {

	appName := info.GetName()
	monitorLabels := labels.GetMonitorLabels(instanceName, appName)
	ls := labels.GetApplicationLabels(appName)

	relabelings := []*servicemonitor.ConfigRelabeling{{
		Action:       "replace",
		SourceLabels: []string{"__meta_kubernetes_pod_node_name"},
		TargetLabel:  "instance",
	}, {
		Action:       "replace",
		SourceLabels: []string{"job"},
		TargetLabel:  "job",
		Replacement:  "caos_remote_${1}",
	}, {
		Action: "labeldrop",
		Regex:  "(container|endpoint|namespace|pod)",
	}}

	metricRelabelings := []*servicemonitor.ConfigRelabeling{{
		Action: "keep",
		SourceLabels: []string{
			"__name__",
			"state",
			"name",
		},
		Regex: `(systemd_unit_state;active;(docker\.service|firewalld\.service|keepalived\.service|kubelet\.service|nginx\.service|node-agentd\.service|sshd\.service))`,
	}, {
		Action:       "replace",
		SourceLabels: []string{"__name__"},
		TargetLabel:  "__name__",
		Regex:        "systemd_unit_state",
		Replacement:  "dist_systemd_unit_active",
	}, {
		Action: "labelkeep",
		Regex:  "__.+|job|instance|name",
	}}

	endpoint := &servicemonitor.ConfigEndpoint{
		Port:              "metrics",
		Path:              "/metrics",
		HonorLabels:       false,
		Relabelings:       relabelings,
		MetricRelabelings: metricRelabelings,
	}

	return &servicemonitor.Config{
		Name:                  "prometheus-ingestion-systemd-servicemonitor",
		Endpoints:             []*servicemonitor.ConfigEndpoint{endpoint},
		JobName:               fmt.Sprintf("ingestion-%s", appName),
		MonitorMatchingLabels: monitorLabels,
		ServiceMatchingLabels: ls,
	}
}
