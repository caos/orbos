package main

import (
	"context"
	"time"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
)

var _ testFunc = bootstrap

func bootstrap(settings programSettings, _ *kubernetes.Spec) interactFunc {
	return func(step uint8, orbctl newOrbctlCommandFunc) (time.Duration, error) {

		timeout := 30 * time.Minute
		bootstrapCtx, bootstrapCtxCancel := context.WithTimeout(settings.ctx, timeout)
		defer bootstrapCtxCancel()

		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		started := time.Now()
		go func() {
			for {
				select {
				case <-ticker.C:
					printProgress(settings, step, started, timeout)
				case <-bootstrapCtx.Done():
					return
				}
			}
		}()

		return 15 * time.Minute, runCommand(settings, orbctl(bootstrapCtx), "--gitops takeoff", true, nil, nil)
	}
}
