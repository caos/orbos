package main

import (
	"strings"
	"time"

	"github.com/afiskon/promtail-client/promtail"
)

func destroyTestFunc(logger promtail.Client) testFunc {
	return func(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) error {

		cmd, err := orbctl()
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
		return simpleRunCommand(cmd, time.NewTimer(5*time.Minute), func(line string) bool {
			logger.Infof(line)
			if !confirmed && strings.HasPrefix(line, "Are you absolutely sure") {
				confirmed = true
				if _, err := stdin.Write([]byte("y\n")); err != nil {
					panic(err)
				}
			}
			return true
		})
	}
}
