package logs

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/logging"
)

func GetFlow(outputs []string, clusterOutputs []string) *logging.FlowConfig {
	ls := map[string]string{
		"app.kubernetes.io/instance": "grafana",
		"app.kubernetes.io/name":     "grafana",
	}

	return &logging.FlowConfig{
		Name:           "flow-grafana",
		Namespace:      "caos-system",
		SelectLabels:   ls,
		Outputs:        outputs,
		ClusterOutputs: clusterOutputs,
		ParserType:     "logfmt",
	}

}
