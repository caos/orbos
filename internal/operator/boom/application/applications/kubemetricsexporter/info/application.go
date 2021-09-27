package info

import "github.com/caos/orbos/v5/internal/operator/boom/name"

const (
	applicationName name.Application = "kube-state-metrics"
	namespace       string           = "caos-system"
)

func GetName() name.Application {
	return applicationName
}

func GetNamespace() string {
	return namespace
}
