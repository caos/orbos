package database

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/metricspersisting/servicemonitor"
	deprecatedlabels "github.com/caos/orbos/internal/operator/boom/labels"
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

		MonitorMatchingLabels: deprecatedlabels.GetMonitorLabels(instanceName, monitorName),

		ServiceMatchingLabels: getApplicationServiceLabels(),
		JobName:               monitorName.String(),
		NamespaceSelector:     []string{"caos-zitadel"},
	}
}

func getOperatorServiceMonitor(instanceName string) *servicemonitor.Config {
	var monitorName name.Application = "database-operator-servicemonitor"

	return &servicemonitor.Config{
		Name: monitorName.String(),
		Endpoints: []*servicemonitor.ConfigEndpoint{{
			Port: "http",
		}},
		MonitorMatchingLabels: deprecatedlabels.GetMonitorLabels(instanceName, monitorName),
		ServiceMatchingLabels: getOperatorServiceLabels(),
		JobName:               monitorName.String(),
		NamespaceSelector:     []string{"caos-system"},
	}
}
