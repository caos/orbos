package main

import (
	"context"
	"strings"
	"time"
)

var _ testFunc = destroyTestFunc

func destroyTestFunc(settings programSettings, orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc, _ uint8) error {

	destroyCtx, destroyCtxCancel := context.WithTimeout(settings.ctx, 5*time.Minute)
	defer destroyCtxCancel()

	cmd := orbctl(destroyCtx)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	var confirmed bool

	return runCommand(settings, cmd, "--gitops destroy", true, nil, func(line string) {
		if !confirmed && strings.HasPrefix(line, "Are you absolutely sure") {
			confirmed = true
			if _, err := stdin.Write([]byte("y\n")); err != nil {
				panic(err)
			}
		}
	})
}
