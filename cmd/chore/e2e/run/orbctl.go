package main

import (
	"os"
	"os/exec"

	"github.com/caos/orbos/cmd/chore"
)

type newOrbctlCommandFunc func() (*exec.Cmd, error)

func buildOrbctl(orbconfig string) (newOrbctlCommandFunc, error) {
	newCmd, err := chore.Orbctl(false)
	if err != nil {
		return nil, err
	}

	version := newCmd()
	version.Args = append(version.Args, "--version")
	version.Stdout = os.Stdout
	version.Stderr = os.Stderr

	if err := version.Run(); err != nil {
		return nil, err
	}

	return func() (*exec.Cmd, error) {
		cmd := newCmd()
		cmd.Args = append(cmd.Args, "--orbconfig", orbconfig)
		return cmd, nil
	}, nil
}
