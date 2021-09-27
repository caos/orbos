package orbiter

import (
	"github.com/caos/orbos/v5/internal/operator/boom/application/applications/metricspersisting/servicemonitor"
	deprecatedlabels "github.com/caos/orbos/v5/internal/operator/boom/labels"
	"github.com/caos/orbos/v5/internal/operator/boom/name"
	"github.com/caos/orbos/v5/internal/operator/orbiter/kinds/orb"
	"github.com/caos/orbos/v5/pkg/labels"
)

func GetServicemonitor(instanceName string) *servicemonitor.Config {
	var appName name.Application
	appName = "orbiter"

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
		MonitorMatchingLabels: deprecatedlabels.GetMonitorLabels(instanceName, appName),
		ServiceMatchingLabels: labels.MustK8sMap(orb.OperatorSelector()),
		JobName:               "orbiter",
	}
}
