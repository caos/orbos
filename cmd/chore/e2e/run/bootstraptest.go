package main

import (
	"context"
	"time"

	"github.com/afiskon/promtail-client/promtail"
)

func bootstrapTestFunc(ctx context.Context, logger promtail.Client, orb string, step uint8) testFunc {
	return func(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) (err error) {

		timeout := 20 * time.Minute
		bootstrapCtx, bootstrapCtxCancel := context.WithTimeout(ctx, timeout)
		defer bootstrapCtxCancel()

		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()

		started := time.Now()
		go func() {
			for {
				select {
				case <-ticker.C:
					printProgress(logger, orb, step, started, timeout)
				case <-bootstrapCtx.Done():
					return
				}
			}
		}()

		return runCommand(logger, orbctl(bootstrapCtx), "--gitops takeoff", true, nil, nil)
	}
}
