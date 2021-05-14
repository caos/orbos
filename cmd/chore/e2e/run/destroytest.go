package main

import (
	"context"
	"strings"
	"time"

	"github.com/afiskon/promtail-client/promtail"
)

func destroyTestFunc(ctx context.Context, logger promtail.Client) testFunc {
	return func(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) error {

		destroyCtx, destroyCtxCancel := context.WithTimeout(ctx, 5*time.Minute)
		defer destroyCtxCancel()

		cmd, err := orbctl(destroyCtx)
		if err != nil {
			return err
		}

		cmd.Args = append(cmd.Args, "--gitops", "destroy")

		errWriter, errWrite := logWriter(logger.Errorf)
		defer errWrite()
		cmd.Stderr = errWriter

		stdin, err := cmd.StdinPipe()
		if err != nil {
			panic(err)
		}

		var confirmed bool

		return simpleRunCommand(cmd, func(line string) {
			logORBITERStdout(logger, line)
			if !confirmed && strings.HasPrefix(line, "Are you absolutely sure") {
				confirmed = true
				if _, err := stdin.Write([]byte("y\n")); err != nil {
					panic(err)
				}
			}
		})
	}
}
