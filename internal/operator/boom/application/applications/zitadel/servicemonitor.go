package zitadel

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/prometheus/servicemonitor"
	deprecatedlabels "github.com/caos/orbos/internal/operator/boom/labels"
	"github.com/caos/orbos/internal/operator/boom/name"
	"github.com/caos/orbos/internal/operator/database/kinds/databases/managed"
	"github.com/caos/orbos/pkg/labels"
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
		MonitorMatchingLabels: deprecatedlabels.GetMonitorLabels(instanceName, monitorName),

		ServiceMatchingLabels: labels.MustK8sMap(managed.PublicServiceNameSelector()),
		JobName:               monitorName.String(),
		NamespaceSelector:     []string{"caos-zitadel"},
	}
}
