package main

import (
	"fmt"
	"os"
	"time"
)

func patchTestFunc(path, value string) func(newOrbctlCommandFunc, newKubectlCommandFunc) error {
	return func(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) error {

		cmd, err := orbctl()
		if err != nil {
			return err
		}

		cmd.Args = append(cmd.Args, "file", "patch", "orbiter.yml", path, "--value", value, "--exact")
		cmd.Stderr = os.Stderr

		return simpleRunCommand(cmd, time.NewTimer(15*time.Second), func(line string) bool {
			fmt.Println(line)
			return true
		})
	}
}
