package database

import "github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/logging"

func GetFlows(outputs []string) []*logging.FlowConfig {
	return []*logging.FlowConfig{
		getOperatorFlow(outputs),
		getDatabaseFlow(outputs),
	}
}

func getOperatorFlow(outputs []string) *logging.FlowConfig {
	return &logging.FlowConfig{
		Name:         "flow-database-operator",
		Namespace:    "caos-system",
		SelectLabels: getOperatorServiceLabels(),
		Outputs:      outputs,
		ParserType:   "logfmt",
	}
}

func getDatabaseFlow(outputs []string) *logging.FlowConfig {
	return &logging.FlowConfig{
		Name:         "flow-database",
		Namespace:    "caos-system",
		SelectLabels: getApplicationServiceLabels(),
		Outputs:      outputs,
		ParserType:   "logfmt",
	}
}
