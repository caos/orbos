package main

import (
	"time"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
)

var _ testFunc = rebootTestFunc

func rebootTestFunc(settings programSettings, spec *kubernetes.Spec) interactFunc {
	return func(_ uint8, orbctl newOrbctlCommandFunc) (time.Duration, error) {

		return 0, nil
		/*

			nodeIDCtx, nodeIDCancel := context.WithTimeout(settings.ctx, 30*time.Second)
			defer nodeIDCancel()

			var nodeID string
			return runCommand(settings, orbctl(nodeIDCtx), fmt.Sprintf("--gitops nodes list --context %s.management --column id", settings.orbID), false, nil, func(line string) {
				nodeID = line
			})*/
	}
}
