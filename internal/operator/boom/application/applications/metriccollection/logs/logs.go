package logs

import (
	"github.com/caos/orbos/v5/internal/operator/boom/application/applications/logcollection/logging"
)

func GetFlow(outputs []string, clusterOutputs []string) *logging.FlowConfig {
	ls := map[string]string{
		"release": "prometheus-operator",
		"app":     "prometheus-operator-operator",
	}

	return &logging.FlowConfig{
		Name:           "flow-prometheus-operator",
		Namespace:      "caos-system",
		SelectLabels:   ls,
		Outputs:        outputs,
		ClusterOutputs: clusterOutputs,
		ParserType:     "logfmt",
	}
}
