package orbiter

import "github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/logging"

func GetFlow(outputs []string) *logging.FlowConfig {
	ls := map[string]string{
		"app.kubernetes.io/part-of":    "ORBOS",
		"app.kubernetes.io/managed-by": "orbiter.caos.ch",
		"app.kubernetes.io/component":  "operator",
	}

	return &logging.FlowConfig{
		Name:         "flow-orbiter",
		Namespace:    "caos-system",
		SelectLabels: ls,
		Outputs:      outputs,
		ParserType:   "logfmt",
	}
}
