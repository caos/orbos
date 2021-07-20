package logs

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/logcollection/logging"
)

func GetFlow(outputs []string, clusterOutputs []string) *logging.FlowConfig {
	ls := map[string]string{
		"app":        "prometheus",
		"prometheus": "caos-prometheus",
	}

	return &logging.FlowConfig{
		Name:           "flow-prometheus",
		Namespace:      "caos-system",
		SelectLabels:   ls,
		Outputs:        outputs,
		ClusterOutputs: clusterOutputs,
		ParserType:     "none",
	}
}
