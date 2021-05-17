package main

import (
	"context"
	"time"
)

var _ testFunc = bootstrapTestFunc

func bootstrapTestFunc(settings programSettings, orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc, step uint8) (err error) {

	timeout := 20 * time.Minute
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

	return runCommand(settings, orbctl(bootstrapCtx), "--gitops takeoff", true, nil, nil)
}
