package main

import (
	"context"
	"fmt"
	"time"
)

func patchTestFunc(path, value string) testFunc {
	return func(settings programSettings, orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc, _ uint8) error {

		patchCtx, patchCancel := context.WithTimeout(settings.ctx, 30*time.Second)
		defer patchCancel()

		cmd := orbctl(patchCtx)
		cmd.Args = append(cmd.Args, "--gitops", "file", "patch", "orbiter.yml", path, "--value", value, "--exact")

		errWriter, errWrite := logWriter(settings.logger.Errorf)
		defer errWrite()
		cmd.Stderr = errWriter

		return runCommand(settings, orbctl(patchCtx), fmt.Sprintf("--gitops file patch orbiter.yml %s --value %s --exact", path, value), true, nil, nil)
	}
}
