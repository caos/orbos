package zitadel

import "github.com/caos/orbos/v5/pkg/labels"

func getOperatorServiceLabels() map[string]string {
	return labels.MustK8sMap(labels.DeriveComponentSelector(labels.MustForComponent(labels.NoopAPI(labels.SelectorOperator("ZITADEL", "zitadel.caos.ch")), "operator"), false))
}

func getApplicationServiceLabels() map[string]string {
	return labels.MustK8sMap(labels.DeriveComponentSelector(labels.MustForComponent(labels.NoopAPI(labels.SelectorOperator("ZITADEL", "zitadel.caos.ch")), "ZITADEL"), false))
}
