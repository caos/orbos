package main

import (
	"time"

	"github.com/afiskon/promtail-client/promtail"
)

func bootstrapTestFunc(logger promtail.Client, timeout time.Duration, step uint8) testFunc {
	return func(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) (err error) {

		cmd, err := orbctl()
		if err != nil {
			return err
		}

		cmd.Args = append(cmd.Args, "--gitops", "takeoff")

		errWriter, errWrite := logWriter(logger.Errorf)
		defer errWrite()
		cmd.Stderr = errWriter

		ticker := time.NewTicker(time.Minute)
		defer ticker.Stop()
		timer := time.NewTimer(timeout)
		defer timer.Stop()

		started := time.Now()
		goon := true
		go func() {
			for {
				select {
				case <-ticker.C:
					printProgress(logger, step, started, timeout)
				case <-timer.C:
					goon = false
				}
			}
		}()

		return simpleRunCommand(cmd, time.NewTimer(20*time.Minute), func(line string) bool {
			logORBITERStdout(logger, line)
			return goon
		})
	}
}
