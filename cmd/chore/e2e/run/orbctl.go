package main

import (
	"bufio"
	"os/exec"

	"github.com/caos/orbos/cmd/chore"
)

type newOrbctlCommandFunc func() (*exec.Cmd, error)

func buildOrbctl(orbconfig string) (newOrbctlCommandFunc, error) {
	newCmd, err := chore.Orbctl(false)
	if err != nil {
		return nil, err
	}

	return func() (*exec.Cmd, error) {
		cmd := newCmd()
		cmd.Args = append(cmd.Args, "--orbconfig", orbconfig)
		return cmd, nil
	}, nil
}

func simpleRunCommand(cmd *exec.Cmd, scan func(line string) bool) error {
	out, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		if !scan(scanner.Text()) {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}
