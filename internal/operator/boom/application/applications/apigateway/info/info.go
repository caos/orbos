package info

import (
	"github.com/caos/orbos/internal/operator/boom/name"
)

const (
	applicationName name.Application = "ambassador"
	namespace       string           = "caos-system"
)

func GetName() name.Application {
	return applicationName
}

func GetNamespace() string {
	return namespace
}
