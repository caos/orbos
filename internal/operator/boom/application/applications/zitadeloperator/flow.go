package zitadeloperator

import "github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/logging"

func GetFlow(outputs []string) *logging.FlowConfig {
	ls := map[string]string{
		"app.kubernetes.io/part-of":   "orbos",
		"app.kubernetes.io/component": "zitadel-operator",
	}

	return &logging.FlowConfig{
		Name:         "flow-zitadel-operator",
		Namespace:    "caos-system",
		SelectLabels: ls,
		Outputs:      outputs,
		ParserType:   "logfmt",
	}
}
