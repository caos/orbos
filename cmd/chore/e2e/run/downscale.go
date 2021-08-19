package main

import (
	"context"
	"fmt"
	"time"
)

var _ testFunc = downscale

func downscale(_ *testSpecs, settings programSettings, conditions *conditions) interactFunc {

	// assignments must be done also when test is skipped
	conditions.kubernetes.ControlPlane.Nodes = 1
	conditions.kubernetes.Workers[0].Nodes = 1
	conditions.orbiter.watcher = watch(10*time.Minute, orbiterPrefix)
	conditions.testCase = nil

	return func(ctx context.Context, _ uint8, orbctl newOrbctlCommandFunc) error {

		if err := patch(ctx, settings, orbctl, fmt.Sprintf("clusters.%s.spec.controlplane.nodes", settings.clusterkey), "1"); err != nil {
			return err
		}

		return patch(ctx, settings, orbctl, fmt.Sprintf("clusters.%s.spec.workers.0.nodes", settings.clusterkey), "1")
	}
}
