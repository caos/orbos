package info

import "github.com/caos/orbos/internal/operator/boom/name"

const (
	applicationName name.Application = "loki"
	orderNumber     int              = 2
	namespace       string           = "caos-system"
)

func GetName() name.Application {
	return applicationName
}

func GetNamespace() string {
	return namespace
}

func GetOrderNumber() int {
	return orderNumber
}
