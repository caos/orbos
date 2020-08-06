package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func run(orbconfig string) error {
	orbctl := curryRunOrbctl(orbconfig)
	return destroy(orbctl)
}

type orbctlFunc func() (*exec.Cmd, func(stdout func(line string) bool) error, error)

func curryRunOrbctl(orbconfig string) orbctlFunc {
	return func() (cmd *exec.Cmd, f func(stdout func(line string) bool) error, err error) {
		return runOrbctl(orbconfig)
	}
}

func destroy(orbctl orbctlFunc) error {

	cmd, run, err := orbctl()
	if err != nil {
		return err
	}

	cmd.Args = append(cmd.Args, "destroy")
	cmd.Stderr = os.Stderr
	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	return run(func(line string) bool {
		fmt.Println(line)
		if strings.HasPrefix(line, "Are you absolutely sure") {
			if _, err := stdin.Write([]byte("y\n")); err != nil {
				panic(err)
			}
		}
		return true
	})
}
