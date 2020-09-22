package boom

import "github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/logging"

func GetFlow(outputs []string) *logging.FlowConfig {
	ls := map[string]string{
		"app.kubernetes.io/part-of":   "orbos",
		"app.kubernetes.io/component": "boom",
	}

	return &logging.FlowConfig{
		Name:         "flow-boom",
		Namespace:    "caos-system",
		SelectLabels: ls,
		Outputs:      outputs,
		ParserType:   "logfmt",
	}
}
