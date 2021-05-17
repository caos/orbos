package main

import (
	"context"
	"time"
)

func patchTestFunc() {

}

func patch(settings programSettings, orbctl newOrbctlCommandFunc, path, value string) error {

	try := func() error {

		patchCtx, patchCancel := context.WithTimeout(settings.ctx, 30*time.Second)
		defer patchCancel()

		cmd := orbctl(patchCtx)
		cmd.Args = append(cmd.Args, "--gitops", "file", "patch", "orbiter.yml", path, "--value", value, "--exact")

		errWriter, errWrite := logWriter(settings.logger.Errorf)
		defer errWrite()
		cmd.Stderr = errWriter

		return runCommand(settings, true, nil, nil, orbctl(patchCtx), "--gitops", "file", "patch", "orbiter.yml", path, "--value", value, "--exact")
	}

	return retry(3, try)
}
