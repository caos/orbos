package database

import "github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/logging"

func GetFlows(outputs []string) []*logging.FlowConfig {
	return []*logging.FlowConfig{
		getOperatorFlow(outputs),
		getDatabaseFlow(outputs),
	}
}

func getOperatorFlow(outputs []string) *logging.FlowConfig {
	ls := map[string]string{
		"app.kubernetes.io/component":  "operator",
		"app.kubernetes.io/managed-by": "database.caos.ch",
		"app.kubernetes.io/part-of":    "ORBOS",
	}

	return &logging.FlowConfig{
		Name:         "flow-database-operator",
		Namespace:    "caos-system",
		SelectLabels: ls,
		Outputs:      outputs,
		ParserType:   "logfmt",
	}
}

func getDatabaseFlow(outputs []string) *logging.FlowConfig {
	ls := map[string]string{
		"app.kubernetes.io/component": "database",
		"app.kubernetes.io/part-of":   "ORBOS",
	}

	return &logging.FlowConfig{
		Name:         "flow-database",
		Namespace:    "caos-system",
		SelectLabels: ls,
		Outputs:      outputs,
		ParserType:   "logfmt",
	}
}
