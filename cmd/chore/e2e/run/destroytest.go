package main

import (
	"fmt"
	"os"
	"strings"
	"time"
)

func destroyTest(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) error {

	cmd, err := orbctl()
	if err != nil {
		return err
	}

	cmd.Args = append(cmd.Args, "--gitops", "destroy")
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	var confirmed bool
	return simpleRunCommand(cmd, time.NewTimer(5*time.Minute), func(line string) bool {
		fmt.Println(line)
		if !confirmed && strings.HasPrefix(line, "Are you absolutely sure") {
			confirmed = true
			if _, err := stdin.Write([]byte("y\n")); err != nil {
				panic(err)
			}
		}
		return true
	})
}
