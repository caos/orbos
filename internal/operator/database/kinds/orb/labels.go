package orb

import "github.com/caos/orbos/pkg/labels"

func mustDatabaseOperator(binaryVersion string) *labels.Operator {
	return labels.MustForOperator("ORBOS", "database.caos.ch", binaryVersion)
}
