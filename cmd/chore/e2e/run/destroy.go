package main

import (
	"context"
	"strings"
	"time"

	"github.com/caos/orbos/internal/operator/orbiter/kinds/clusters/kubernetes"
)

var _ testFunc = destroy

func destroy(settings programSettings, _ *kubernetes.Spec) interactFunc {

	return func(_ uint8, orbctl newOrbctlCommandFunc) (time.Duration, error) {

		try := func() error {

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

		return 0, retry(3, try)
	}
}
