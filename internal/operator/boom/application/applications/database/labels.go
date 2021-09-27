package database

import (
	"github.com/caos/orbos/v5/pkg/labels"
)

func getOperatorServiceLabels() map[string]string {
	return labels.MustK8sMap(labels.DeriveComponentSelector(labels.MustForComponent(labels.NoopAPI(labels.SelectorOperator("ZITADEL", "database.caos.ch")), "operator"), false))
}

func getApplicationServiceLabels() map[string]string {
	return labels.MustK8sMap(labels.DeriveComponentSelector(labels.MustForComponent(labels.NoopAPI(labels.SelectorOperator("ZITADEL", "database.caos.ch")), "database"), false))
}
