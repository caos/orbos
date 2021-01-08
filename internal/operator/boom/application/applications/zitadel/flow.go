package zitadel

import "github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/logging"

func GetFlows(outputs []string) []*logging.FlowConfig {
	return []*logging.FlowConfig{
		getOperatorFlow(outputs),
		getZitadelFlow(outputs),
	}
}

func getZitadelFlow(outputs []string) *logging.FlowConfig {
	return &logging.FlowConfig{
		Name:         "flow-zitadel",
		Namespace:    "caos-system",
		SelectLabels: getApplicationServiceLabels(),
		Outputs:      outputs,
		ParserType:   "logfmt",
	}
}

func getOperatorFlow(outputs []string) *logging.FlowConfig {
	return &logging.FlowConfig{
		Name:         "flow-zitadel-operator",
		Namespace:    "caos-system",
		SelectLabels: getOperatorServiceLabels(),
		Outputs:      outputs,
		ParserType:   "logfmt",
	}
}
