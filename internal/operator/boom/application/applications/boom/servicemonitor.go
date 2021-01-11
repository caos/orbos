package boom

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/servicemonitor"
	deprecatedlabels "github.com/caos/orbos/internal/operator/boom/labels"
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/pkg/labels"
)

func GetServicemonitor(instanceName string) *servicemonitor.Config {
	var appName name.Application
	appName = "boom"

	endpoint := &servicemonitor.ConfigEndpoint{
		Port: "metrics",
		Path: "/metrics",
	}

	return &servicemonitor.Config{
		Name:                  "boom-servicemonitor",
		Endpoints:             []*servicemonitor.ConfigEndpoint{endpoint},
		MonitorMatchingLabels: deprecatedlabels.GetMonitorLabels(instanceName, appName),
		ServiceMatchingLabels: labels.MustK8sMap(labels.OpenOperatorSelector("boom.caos.ch")),
		JobName:               "boom",
	}
}
