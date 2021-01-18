package database

import (
	"github.com/caos/orbos/internal/operator/database/kinds/databases"
	"github.com/caos/orbos/internal/operator/database/kinds/orb"
	"github.com/caos/orbos/pkg/labels"
)

func getOperatorServiceLabels() map[string]string {
	return labels.MustK8sMap(orb.OperatorSelector())
}

func getApplicationServiceLabels() map[string]string {
	return labels.MustK8sMap(databases.ComponentSelector())
}
