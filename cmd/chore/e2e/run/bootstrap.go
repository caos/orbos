package main

import (
	"context"
	"fmt"
	"time"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
)

var _ testFunc = bootstrap

func bootstrap(settings programSettings, _ *kubernetes.Spec) interactFunc {
	return func(step uint8, orbctl newOrbctlCommandFunc) (time.Duration, checkCurrentFunc, error) {

		timeout := 15 * time.Minute
		bootstrapCtx, bootstrapCtxCancel := context.WithTimeout(settings.ctx, timeout)
		defer bootstrapCtxCancel()

		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		started := time.Now()
		go func() {
			for {
				select {
				case <-ticker.C:
					printProgress(settings, fmt.Sprintf("%d (takeoff)", step), started, timeout)
				case <-bootstrapCtx.Done():
					return
				}
			}
		}()

		return 15 * time.Minute, nil, runCommand(settings, true, nil, nil, orbctl(bootstrapCtx), "--gitops", "takeoff")
	}
}
