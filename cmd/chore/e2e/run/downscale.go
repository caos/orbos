package main

import (
	"fmt"
	"time"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
)

var _ testFunc = downscale

func downscale(settings programSettings, expect *kubernetes.Spec) interactFunc {
	expect.ControlPlane.Nodes = 1
	expect.Workers[0].Nodes = 2
	return func(_ uint8, orbctl newOrbctlCommandFunc) (time.Duration, checkCurrentFunc, error) {

		if err := patch(settings, orbctl, fmt.Sprintf("clusters.%s.spec.controlplane.nodes", settings.orbID), "1"); err != nil {
			return 0, nil, err
		}

		if err := patch(settings, orbctl, fmt.Sprintf("clusters.%s.spec.workers.0.nodes", settings.orbID), "2"); err != nil {
			return 0, nil, err
		}

		return 10 * time.Minute, nil, nil
	}
}
