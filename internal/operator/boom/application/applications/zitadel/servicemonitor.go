package zitadel

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/servicemonitor"
	"github.com/caos/orbos/internal/operator/boom/labels"
	"github.com/caos/orbos/internal/operator/boom/name"
)

func GetServicemonitor(instanceName string) *servicemonitor.Config {

	var monitorName name.Application = "zitadel-cockroachdb"

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
			"app.kubernetes.io/component": "iam-database",
			"app.kubernetes.io/part-of":   "zitadel",
			"zitadel.caos.ch/servicetype": "internal",
		},
		JobName:           monitorName.String(),
		NamespaceSelector: []string{"caos-zitadel"},
	}
}
