package networking

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/servicemonitor"
	"github.com/caos/orbos/internal/operator/boom/labels"
	"github.com/caos/orbos/internal/operator/boom/name"
)

func GetServicemonitor(instanceName string) *servicemonitor.Config {
	var monitorName name.Application = "networking-operator-servicemonitor"

	return &servicemonitor.Config{
		Name: monitorName.String(),
		Endpoints: []*servicemonitor.ConfigEndpoint{{
			Port: "http",
		}},
		MonitorMatchingLabels: labels.GetMonitorLabels(instanceName, monitorName),
		ServiceMatchingLabels: map[string]string{
			"app.kubernetes.io/component":  "operator",
			"app.kubernetes.io/managed-by": "networking.caos.ch",
			"app.kubernetes.io/part-of":    "ORBOS",
		},
		JobName:           monitorName.String(),
		NamespaceSelector: []string{"caos-system"},
	}
}
