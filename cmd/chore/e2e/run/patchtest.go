package main

import (
	"context"
	"time"

	"github.com/afiskon/promtail-client/promtail"
)

func patchTestFunc(ctx context.Context, logger promtail.Client, path, value string) func(newOrbctlCommandFunc, newKubectlCommandFunc) error {
	return func(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) error {

		patchCtx, patchCancel := context.WithTimeout(ctx, 30*time.Second)
		defer patchCancel()

		cmd, err := orbctl(patchCtx)
		if err != nil {
			return err
		}

		cmd.Args = append(cmd.Args, "--gitops", "file", "patch", "orbiter.yml", path, "--value", value, "--exact")
		errWriter, errWrite := logWriter(logger.Errorf)
		defer errWrite()
		cmd.Stderr = errWriter

		return simpleRunCommand(cmd, func(line string) {
			logORBITERStdout(logger, line)
		})
	}
}
