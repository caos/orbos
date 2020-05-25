package logs

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/logging"
)

func GetFlow(outputs []string) *logging.FlowConfig {
	ls := map[string]string{
		"release": "prometheus-operator",
		"app":     "prometheus-operator-operator",
	}

	return &logging.FlowConfig{
		Name:         "flow-prometheus-operator",
		Namespace:    "caos-system",
		SelectLabels: ls,
		Outputs:      outputs,
		ParserType:   "logfmt",
	}
}
