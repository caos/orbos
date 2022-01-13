package boom

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/logcollection/logging"
)

func GetFlow(outputs, clusterOutputs []string) *logging.FlowConfig {

	return &logging.FlowConfig{
		Name:           "flow-boom",
		Namespace:      "caos-system",
		SelectLabels:   getOperatorServiceLabels(),
		Outputs:        outputs,
		ClusterOutputs: clusterOutputs,
		ParserType:     "logfmt",
	}
}
