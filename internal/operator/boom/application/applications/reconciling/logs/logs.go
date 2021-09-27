package logs

import (
	"github.com/caos/orbos/v5/internal/operator/boom/application/applications/logcollection/logging"
)

func GetFlow(outputs []string, clusterOutputs []string) *logging.FlowConfig {
	ls := map[string]string{
		"app.kubernetes.io/instance": "argocd",
		"app.kubernetes.io/part-of":  "argocd",
	}

	return &logging.FlowConfig{
		Name:           "flow-argocd",
		Namespace:      "caos-system",
		SelectLabels:   ls,
		Outputs:        outputs,
		ClusterOutputs: clusterOutputs,
		ParserType:     "logfmt",
	}
}
