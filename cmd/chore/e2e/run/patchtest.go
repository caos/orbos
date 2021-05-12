package main

import (
	"time"

	"github.com/afiskon/promtail-client/promtail"
)

func patchTestFunc(logger promtail.Client, path, value string) func(newOrbctlCommandFunc, newKubectlCommandFunc) error {
	return func(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) error {

		cmd, err := orbctl()
		if err != nil {
			return err
		}

		cmd.Args = append(cmd.Args, "--gitops", "file", "patch", "orbiter.yml", path, "--value", value, "--exact")
		errWriter, errWrite := logWriter(logger.Errorf)
		defer errWrite()
		cmd.Stderr = errWriter

		return simpleRunCommand(cmd, time.NewTimer(15*time.Second), func(line string) bool {
			logORBITERStdout(logger, line)
			return true
		})
	}
}
