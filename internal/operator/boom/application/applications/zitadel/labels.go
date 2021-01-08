package zitadel

import "github.com/caos/orbos/pkg/labels"

func getOperatorServiceLabels() map[string]string {
	return labels.MustK8sMap(labels.DeriveComponentSelector(labels.MustForComponent(labels.MustForAPI(labels.MustForOperator("ZITADEL", "zitadel", "v1"), "", ""), "operator"), false))
}

func getApplicationServiceLabels() map[string]string {
	return labels.MustK8sMap(labels.DeriveComponentSelector(labels.MustForComponent(labels.MustForAPI(labels.NoopOperator("ZITADEL"), "", ""), "zitadel"), false))
}
