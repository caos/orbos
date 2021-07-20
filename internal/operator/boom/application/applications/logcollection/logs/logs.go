package logs

import "github.com/caos/orbos/internal/operator/boom/application/applications/logcollection/logging"

func GetFlow(outputs []string, clusterOutputs []string) *logging.FlowConfig {
	lables := map[string]string{"app.kubernetes.io/instance": "logging-operator", "app.kubernetes.io/name": "logging-operator"}

	return &logging.FlowConfig{
		Name:           "flow-logging-operator",
		Namespace:      "caos-system",
		SelectLabels:   lables,
		Outputs:        outputs,
		ClusterOutputs: clusterOutputs,
		ParserType:     "json",
	}
}
