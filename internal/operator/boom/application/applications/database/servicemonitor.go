package database

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/servicemonitor"
	"github.com/caos/orbos/internal/operator/boom/labels"
	"github.com/caos/orbos/internal/operator/boom/name"
)

func GetServicemonitors(instanceName string) []*servicemonitor.Config {
	return []*servicemonitor.Config{
		getDatabaseServiceMonitor(instanceName),
		getOperatorServiceMonitor(instanceName),
	}
}

func getDatabaseServiceMonitor(instanceName string) *servicemonitor.Config {
	var monitorName name.Application = "database-servicemonitor"

	return &servicemonitor.Config{
		Name: monitorName.String(),
		Endpoints: []*servicemonitor.ConfigEndpoint{{
			Port: "http",
			Path: "/_status/vars",
			TLSConfig: &servicemonitor.ConfigTLSConfig{
				InsecureSkipVerify: true,
			},
			Relabelings: []*servicemonitor.ConfigRelabeling{{
				Action:       "replace",
				SourceLabels: []string{"__meta_kubernetes_pod_node_name"},
				TargetLabel:  "instance",
			}},
		}},
		MonitorMatchingLabels: labels.GetMonitorLabels(instanceName, monitorName),
		ServiceMatchingLabels: map[string]string{
			"app.kubernetes.io/component": "cockroachdb",
			"app.kubernetes.io/part-of":   "ORBOS",
		},
		JobName:           monitorName.String(),
		NamespaceSelector: []string{"caos-zitadel"},
	}
}

func getOperatorServiceMonitor(instanceName string) *servicemonitor.Config {
	var monitorName name.Application = "database-operator-servicemonitor"

	return &servicemonitor.Config{
		Name: monitorName.String(),
		Endpoints: []*servicemonitor.ConfigEndpoint{{
			Port: "http",
		}},
		MonitorMatchingLabels: labels.GetMonitorLabels(instanceName, monitorName),
		ServiceMatchingLabels: map[string]string{
			"app.kubernetes.io/component":  "operator",
			"app.kubernetes.io/managed-by": "database.caos.ch",
			"app.kubernetes.io/part-of":    "ORBOS",
		},
		JobName:           monitorName.String(),
		NamespaceSelector: []string{"caos-system"},
	}
}
