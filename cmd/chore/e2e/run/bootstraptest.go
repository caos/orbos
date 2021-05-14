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

		cmd, err := orbctl(bootstrapCtx)
		if err != nil {
			return err
		}

		cmd.Args = append(cmd.Args, "--gitops", "takeoff")

		errWriter, errWrite := logWriter(logger.Errorf)
		defer errWrite()
		cmd.Stderr = errWriter

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

		return simpleRunCommand(cmd, func(line string) {
			logORBITERStdout(logger, line)
		})
	}
}
