package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"
)

func ensureORBITERTest(timeout time.Duration) func(newOrbctlCommandFunc, newKubectlCommandFunc) error {
	return func(_ newOrbctlCommandFunc, kubectl newKubectlCommandFunc) error {
		return watchLogs(kubectl, time.NewTimer(timeout))
	}
}

func watchLogs(kubectl newKubectlCommandFunc, timer *time.Timer) error {
	cmd := kubectl()
	cmd.Args = append(cmd.Args, "--namespace", "caos-system", "logs", "-f", "-l", "app=orbiter")
	cmd.Stderr = os.Stderr

	err := simpleRunCommand(cmd, timer, func(line string) (goon bool) {
		fmt.Println(line)
		return !strings.Contains(line, "Desired state is ensured")
	})
	if !errors.Is(err, errTimeout) {
		return watchLogs(kubectl, timer)
	}
	return err
}
