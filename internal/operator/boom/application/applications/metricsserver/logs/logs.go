package logs

import "github.com/caos/orbos/v5/internal/operator/boom/application/applications/logcollection/logging"

func GetFlow(outputs []string, clusterOutputs []string) *logging.FlowConfig {
	ls := map[string]string{
		"release": "metrics-server",
		"app":     "metrics-server",
	}

	return &logging.FlowConfig{
		Name:           "flow-metrics-server",
		Namespace:      "caos-system",
		SelectLabels:   ls,
		Outputs:        outputs,
		ClusterOutputs: clusterOutputs,
		ParserType:     "logfmt",
	}
}
