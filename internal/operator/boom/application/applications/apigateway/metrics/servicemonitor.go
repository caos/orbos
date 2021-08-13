package metrics

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/apigateway/info"
	"github.com/caos/orbos/internal/operator/boom/application/applications/metricspersisting/servicemonitor"
	"github.com/caos/orbos/internal/operator/boom/labels"
)

func GetServicemonitor(instanceName string) *servicemonitor.Config {
	appName := info.GetName()
	monitorlabels := labels.GetMonitorLabels(instanceName, appName)
	ls := labels.GetApplicationLabels(appName)

	endpoint := &servicemonitor.ConfigEndpoint{
		Port: "ambassador-admin",
		Path: "/metrics",
	}

	ls["service"] = "ambassador-admin"
	ls["app.kubernetes.io/part-of"] = "ambassador"
	ls["app.kubernetes.io/name"] = "ambassador"
	ls["app.kubernetes.io/instance"] = "ambassador"

	return &servicemonitor.Config{
		Name:                  "ambassador-servicemonitor",
		Endpoints:             []*servicemonitor.ConfigEndpoint{endpoint},
		MonitorMatchingLabels: monitorlabels,
		ServiceMatchingLabels: ls,
		JobName:               "ambassador",
	}
}

func GetCloudServicemonitor(instanceName string) *servicemonitor.Config {
	sm := GetCloudServicemonitor(instanceName)
	sm.Name = "ambassador-cloud-servicemonitor"

	relabel := &servicemonitor.ConfigRelabeling{
		Regex:        "envoy_(.+)",
		SourceLabels: []string{"__name__"},
		TargetLabel:  "__name__",
		Replacement:  "caos_ambassador_$1",
	}

	sm.Endpoints[0].Relabelings = []*servicemonitor.ConfigRelabeling{relabel}

	return sm
}
