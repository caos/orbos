package boom

import "github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/logging"

func GetFlow(outputs []string) *logging.FlowConfig {
	ls := map[string]string{
		"app.kubernetes.io/component":  "operator",
		"app.kubernetes.io/managed-by": "boom.caos.ch",
		"app.kubernetes.io/part-of":    "ORBOS",
	}

	return &logging.FlowConfig{
		Name:         "flow-boom",
		Namespace:    "caos-system",
		SelectLabels: ls,
		Outputs:      outputs,
		ParserType:   "logfmt",
	}
}
