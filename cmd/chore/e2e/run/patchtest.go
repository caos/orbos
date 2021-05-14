package main

import (
	"context"
	"fmt"
	"time"

	"github.com/afiskon/promtail-client/promtail"
)

func patchTestFunc(ctx context.Context, logger promtail.Client, path, value string) func(newOrbctlCommandFunc, newKubectlCommandFunc) error {
	return func(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) error {

		patchCtx, patchCancel := context.WithTimeout(ctx, 30*time.Second)
		defer patchCancel()

		cmd := orbctl(patchCtx)
		cmd.Args = append(cmd.Args, "--gitops", "file", "patch", "orbiter.yml", path, "--value", value, "--exact")

		errWriter, errWrite := logWriter(logger.Errorf)
		defer errWrite()
		cmd.Stderr = errWriter

		return runCommand(logger, orbctl(patchCtx), fmt.Sprintf("--gitops file patch orbiter.yml %s --value %s --exact", path, value), true, nil, nil)
	}
}
