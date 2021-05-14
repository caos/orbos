package main

import (
	"context"
	"fmt"
	"time"

	"github.com/afiskon/promtail-client/promtail"
)

func rebootTestFunc(ctx context.Context, logger promtail.Client, orb string) func(newOrbctlCommandFunc, newKubectlCommandFunc) error {
	return func(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) error {

		nodeIDCtx, nodeIDCancel := context.WithTimeout(ctx, 30*time.Second)
		defer nodeIDCancel()

		var nodeID string
		return runCommand(logger, orbctl(nodeIDCtx), fmt.Sprintf("--gitops nodes list --context %s.management --column id", orb), false, nil, func(line string) {
			nodeID = line
		})
	}
}
