package networking

import (
	"github.com/caos/orbos/v5/internal/operator/boom/application/applications/logcollection/logging"
	"github.com/caos/orbos/v5/internal/operator/networking/kinds/orb"
	"github.com/caos/orbos/v5/pkg/labels"
)

func GetFlow(outputs, clusterOutputs []string) *logging.FlowConfig {
	return &logging.FlowConfig{
		Name:           "flow-networking-operator",
		Namespace:      "caos-system",
		SelectLabels:   labels.MustK8sMap(orb.OperatorSelector()),
		Outputs:        outputs,
		ClusterOutputs: clusterOutputs,
		ParserType:     "logfmt",
	}
}
