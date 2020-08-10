package main

import (
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
