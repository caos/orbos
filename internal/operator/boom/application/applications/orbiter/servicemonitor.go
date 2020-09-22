package orbiter

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/servicemonitor"
	"github.com/caos/orbos/internal/operator/boom/labels"
	"github.com/caos/orbos/internal/operator/boom/name"
)

func GetServicemonitor(instanceName string) *servicemonitor.Config {
	var appName name.Application
	appName = "orbiter"
	monitorlabels := labels.GetMonitorLabels(instanceName, appName)
	ls := map[string]string{
		"app.kubernetes.io/component": "orbiter",
		"app.kubernetes.io/part-of":   "orbos",
	}

	relabelings := []*servicemonitor.ConfigRelabeling{{
		Action:       "replace",
		SourceLabels: []string{"job"},
		TargetLabel:  "job",
		Replacement:  "caos_remote_${1}",
	}, {
		Action: "labeldrop",
		Regex:  "(container|endpoint|namespace|pod)",
	}}

	metricRelabelings := []*servicemonitor.ConfigRelabeling{{
		Action:       "keep",
		Regex:        "probe",
		SourceLabels: []string{"__name__"},
	}, {
		Action: "labelkeep",
		Regex:  "__.+|job|name|type|target",
	}, {
		Action:       "replace",
		SourceLabels: []string{"__name__"},
		TargetLabel:  "__name__",
		Replacement:  "caos_${1}",
	}}

	endpoint := &servicemonitor.ConfigEndpoint{
		Port:              "metrics",
		Path:              "/metrics",
		Relabelings:       relabelings,
		MetricRelabelings: metricRelabelings,
	}

	return &servicemonitor.Config{
		Name:                  "orbiter-servicemonitor",
		Endpoints:             []*servicemonitor.ConfigEndpoint{endpoint},
		MonitorMatchingLabels: monitorlabels,
		ServiceMatchingLabels: ls,
		JobName:               "orbiter",
	}
}
