package info

import "github.com/caos/orbos/v5/internal/operator/boom/name"

const (
	applicationName name.Application = "metrics-server"
	namespace       string           = "kube-system"
)

func GetName() name.Application {
	return applicationName
}

func GetNamespace() string {
	return namespace
}
