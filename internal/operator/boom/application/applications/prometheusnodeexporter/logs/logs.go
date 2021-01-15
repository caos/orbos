package logs

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/logging"
)

func GetFlow(outputs []string, clusterOutputs []string) *logging.FlowConfig {
	ls := map[string]string{
		"release": "prometheus-node-exporter",
		"app":     "prometheus-node-exporter",
	}

	return &logging.FlowConfig{
		Name:           "flow-prometheus-node-exporter",
		Namespace:      "caos-system",
		SelectLabels:   ls,
		Outputs:        outputs,
		ClusterOutputs: clusterOutputs,
		ParserType:     "logfmt",
	}
}
