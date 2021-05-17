package main

import (
	"context"
	"fmt"
	"time"
)

func rebootTestFunc(settings programSettings, orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) error {

	nodeIDCtx, nodeIDCancel := context.WithTimeout(settings.ctx, 30*time.Second)
	defer nodeIDCancel()

	var nodeID string
	return runCommand(settings, orbctl(nodeIDCtx), fmt.Sprintf("--gitops nodes list --context %s.management --column id", settings.orbID), false, nil, func(line string) {
		nodeID = line
	})
}
