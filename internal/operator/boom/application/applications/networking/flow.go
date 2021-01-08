package networking

import (
	"github.com/caos/orbos/internal/operator/boom/application/applications/loggingoperator/logging"
	"github.com/caos/orbos/internal/operator/networking/kinds/orb"
	"github.com/caos/orbos/pkg/labels"
)

func GetFlow(outputs []string) *logging.FlowConfig {
	return &logging.FlowConfig{
		Name:         "flow-networking-operator",
		Namespace:    "caos-system",
		SelectLabels: labels.MustK8sMap(orb.OperatorSelector()),
		Outputs:      outputs,
		ParserType:   "logfmt",
	}
}
