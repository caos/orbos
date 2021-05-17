package main

import (
	"fmt"
	"time"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
)

var _ testFunc = upgrade

func upgrade(settings programSettings, spec *kubernetes.Spec) interactFunc {

	spec.Versions.Kubernetes = "v1.21.0"

	return func(_ uint8, orbctl newOrbctlCommandFunc) (time.Duration, checkCurrentFunc, error) {

		return 5 * time.Minute, nil, patch(settings, orbctl, fmt.Sprintf("clusters.%s.spec.versions.kubernetes", settings.orbID), "v1.21.0")
	}
}
