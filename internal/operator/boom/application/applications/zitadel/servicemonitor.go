package zitadel

import (
	"github.com/caos/orbos/v5/internal/operator/boom/application/applications/metricspersisting/servicemonitor"
	deprecatedlabels "github.com/caos/orbos/v5/internal/operator/boom/labels"
	"github.com/caos/orbos/v5/internal/operator/boom/name"
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
