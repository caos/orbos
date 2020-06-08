package logs

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/logging"
)

func GetFlow(outputs []string) *logging.FlowConfig {
	ls := map[string]string{
		"app.kubernetes.io/instance": "kube-state-metrics",
		"app.kubernetes.io/name":     "kube-state-metrics",
	}

	return &logging.FlowConfig{
		Name:         "flow-kube-state-metrics",
		Namespace:    "caos-system",
		SelectLabels: ls,
		Outputs:      outputs,
		ParserType:   "none",
	}
}
