package zitadel

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/servicemonitor"
	"github.com/caos/orbos/internal/operator/boom/labels"
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
		MonitorMatchingLabels: labels.GetMonitorLabels(instanceName, monitorName),
		ServiceMatchingLabels: map[string]string{
			"app.kubernetes.io/part-of":   "ZITADEL",
			"app.kubernetes.io/component": "ZITADEL",
		},
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
		MonitorMatchingLabels: labels.GetMonitorLabels(instanceName, monitorName),
		ServiceMatchingLabels: map[string]string{
			"app.kubernetes.io/part-of":    "ZITADEL",
			"app.kubernetes.io/managed-by": "zitadel.caos.ch",
			"app.kubernetes.io/component":  "operator",
		},
		JobName:           monitorName.String(),
		NamespaceSelector: []string{"caos-system"},
	}
}
