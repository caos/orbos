package boom

import "github.com/caos/orbos/pkg/labels"

func getOperatorServiceLabels() map[string]string {
	return labels.MustK8sMap(labels.OpenOperatorSelector("ORBOS", "boom.caos.ch"))
}
