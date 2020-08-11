package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func ensureORBITERTest(timeout time.Duration) func(newOrbctlCommandFunc, newKubectlCommandFunc) error {
	return func(_ newOrbctlCommandFunc, kubectl newKubectlCommandFunc) error {

		cmd := kubectl()
		cmd.Args = append(cmd.Args, "--namespace", "caos-system", "logs", "-f", "-l", "app=orbiter")
		cmd.Stderr = os.Stderr

		return simpleRunCommand(cmd, time.NewTimer(timeout), func(line string) bool {
			fmt.Println(line)
			return strings.Contains(line, "Desired state is ensured")
		})
	}
}
