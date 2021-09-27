package networking

import (
	"github.com/caos/orbos/v5/internal/operator/boom/application/applications/metricspersisting/servicemonitor"
	deprecatedlabels "github.com/caos/orbos/v5/internal/operator/boom/labels"
	"github.com/caos/orbos/v5/internal/operator/boom/name"
	"github.com/caos/orbos/v5/internal/operator/networking/kinds/orb"
	"github.com/caos/orbos/v5/pkg/labels"
)

func GetServicemonitor(instanceName string) *servicemonitor.Config {
	var monitorName name.Application = "networking-operator-servicemonitor"

	return &servicemonitor.Config{
		Name: monitorName.String(),
		Endpoints: []*servicemonitor.ConfigEndpoint{{
			Port: "http",
		}},
		MonitorMatchingLabels: deprecatedlabels.GetMonitorLabels(instanceName, monitorName),
		ServiceMatchingLabels: labels.MustK8sMap(orb.OperatorSelector()),
		JobName:               monitorName.String(),
		NamespaceSelector:     []string{"caos-system"},
	}
}
