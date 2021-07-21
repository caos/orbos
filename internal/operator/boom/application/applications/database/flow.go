package database

import "github.com/caos/orbos/internal/operator/boom/application/applications/logcollection/logging"

func GetFlows(outputs, clusterOutputs []string) []*logging.FlowConfig {
	return []*logging.FlowConfig{
		getOperatorFlow(outputs, clusterOutputs),
		getDatabaseFlow(outputs, clusterOutputs),
	}
}

func getOperatorFlow(outputs, clusterOutputs []string) *logging.FlowConfig {
	return &logging.FlowConfig{
		Name:           "flow-database-operator",
		Namespace:      "caos-system",
		SelectLabels:   getOperatorServiceLabels(),
		Outputs:        outputs,
		ClusterOutputs: clusterOutputs,
		ParserType:     "logfmt",
	}
}

func getDatabaseFlow(outputs, clusterOutputs []string) *logging.FlowConfig {
	return &logging.FlowConfig{
		Name:           "flow-database",
		Namespace:      "caos-system",
		SelectLabels:   getApplicationServiceLabels(),
		Outputs:        outputs,
		ClusterOutputs: clusterOutputs,
		ParserType:     "logfmt",
	}
}
