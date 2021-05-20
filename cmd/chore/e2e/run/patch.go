package main

import (
	"context"
	"time"
)

func patch(settings programSettings, orbctl newOrbctlCommandFunc, path, value string) error {

	patchCtx, patchCancel := context.WithTimeout(settings.ctx, 30*time.Second)
	defer patchCancel()

	cmd := orbctl(patchCtx)
	cmd.Args = append(cmd.Args, "--gitops", "file", "patch", "orbiter.yml", path, "--value", value, "--exact")

	return runCommand(settings, true, nil, nil, orbctl(patchCtx), "--gitops", "file", "patch", "orbiter.yml", path, "--value", value, "--exact")
}
