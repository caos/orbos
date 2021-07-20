package metrics

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/metricspersisting/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/metricspersisting/servicemonitor"
	"github.com/caos/orbos/internal/operator/boom/labels"
)

func GetServicemonitor(instanceName string) *servicemonitor.Config {
	appName := info.GetName()
	monitorlabels := labels.GetMonitorLabels(instanceName, appName)
	ls := labels.GetApplicationLabels(appName)

	endpoint := &servicemonitor.ConfigEndpoint{
		Port: "web",
		Path: "/metrics",
	}

	return &servicemonitor.Config{
		Name:                  "prometheus-servicemonitor",
		Endpoints:             []*servicemonitor.ConfigEndpoint{endpoint},
		MonitorMatchingLabels: monitorlabels,
		ServiceMatchingLabels: ls,
		JobName:               appName.String(),
	}
}
