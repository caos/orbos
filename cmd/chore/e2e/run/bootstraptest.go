package main

import (
	"fmt"
	"os"
	"time"
)

func bootstrapTest(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) (err error) {

	cmd, err := orbctl()
	if err != nil {
		return err
	}

	cmd.Args = append(cmd.Args, "takeoff")
	cmd.Stderr = os.Stderr

	return simpleRunCommand(cmd, time.NewTimer(20*time.Minute), func(line string) bool {
		fmt.Println(line)
		return true
	})
}
