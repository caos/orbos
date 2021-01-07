package zitadel

import "github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/logging"

func GetFlows(outputs []string) []*logging.FlowConfig {
	return []*logging.FlowConfig{
		getOperatorFlow(outputs),
		getZitadelFlow(outputs),
	}
}

func getZitadelFlow(outputs []string) *logging.FlowConfig {
	ls := map[string]string{
		"app.kubernetes.io/part-of":   "ZITADEL",
		"app.kubernetes.io/component": "operator",
	}

	return &logging.FlowConfig{
		Name:         "flow-zitadel",
		Namespace:    "caos-system",
		SelectLabels: ls,
		Outputs:      outputs,
		ParserType:   "logfmt",
	}
}

func getOperatorFlow(outputs []string) *logging.FlowConfig {
	ls := map[string]string{
		"app.kubernetes.io/part-of":    "ZITADEL",
		"app.kubernetes.io/managed-by": "zitadel.caos.ch",
		"app.kubernetes.io/component":  "operator",
	}

	return &logging.FlowConfig{
		Name:         "flow-zitadel",
		Namespace:    "caos-system",
		SelectLabels: ls,
		Outputs:      outputs,
		ParserType:   "logfmt",
	}
}
