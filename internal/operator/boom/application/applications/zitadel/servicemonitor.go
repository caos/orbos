package zitadel

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/metricspersisting/servicemonitor"
	deprecatedlabels "github.com/caos/orbos/internal/operator/boom/labels"
	"github.com/caos/orbos/internal/operator/boom/name"
)

func GetServicemonitors(instanceName string) []*servicemonitor.Config {
	return []*servicemonitor.Config{
		getOperatorServicemonitor(instanceName),
		getZitadelServicemonitor(instanceName),
	}
}

func getZitadelServicemonitor(instanceName string) *servicemonitor.Config {
	var monitorName name.Application = "zitadel-servicemonitor"

	return &servicemonitor.Config{
		Name: monitorName.String(),
		Endpoints: []*servicemonitor.ConfigEndpoint{{
			Port: "http",
		}},
		MonitorMatchingLabels: deprecatedlabels.GetMonitorLabels(instanceName, monitorName),
		ServiceMatchingLabels: getApplicationServiceLabels(),

		JobName:           monitorName.String(),
		NamespaceSelector: []string{"caos-zitadel"},
	}
}

func GetCloudZitadelServiceMonitor(instanceName string) *servicemonitor.Config {
	sm := getZitadelServicemonitor(instanceName)

	relabel := &servicemonitor.ConfigRelabeling{
		Regex:        "zitadel_(.+)",
		SourceLabels: []string{"__name__"},
		TargetLabel:  "__name__",
		Replacement:  "caos_zitadel_$1",
	}

	sm.Endpoints[0].Relabelings = append(sm.Endpoints[0].Relabelings, relabel)
	sm.Name = "zitadel-cloud-servicemonitor"
	return sm
}

func getOperatorServicemonitor(instanceName string) *servicemonitor.Config {
	var monitorName name.Application = "zitadel-operator-servicemonitor"

	return &servicemonitor.Config{
		Name: monitorName.String(),
		Endpoints: []*servicemonitor.ConfigEndpoint{{
			Port: "http",
		}},
		MonitorMatchingLabels: deprecatedlabels.GetMonitorLabels(instanceName, monitorName),
		ServiceMatchingLabels: getOperatorServiceLabels(),

		JobName:           monitorName.String(),
		NamespaceSelector: []string{"caos-system"},
	}
}
