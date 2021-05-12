package main

import (
	"time"

	"github.com/afiskon/promtail-client/promtail"
)

func bootstrapTestFunc(logger promtail.Client) testFunc {
	return func(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) (err error) {

		cmd, err := orbctl()
		if err != nil {
			return err
		}

		cmd.Args = append(cmd.Args, "--gitops", "takeoff")

		errWriter, errWrite := logWriter(logger.Errorf)
		defer errWrite()
		cmd.Stderr = errWriter

		return simpleRunCommand(cmd, time.NewTimer(20*time.Minute), func(line string) bool {
			logORBITERStdout(logger, line)
			return true
		})
	}
}
