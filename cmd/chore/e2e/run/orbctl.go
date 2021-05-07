package main

import (
	"os/exec"

	"github.com/afiskon/promtail-client/promtail"

	"github.com/caos/orbos/cmd/chore"
)

type newOrbctlCommandFunc func() (*exec.Cmd, error)

func buildOrbctl(logger promtail.Client, orbconfig string) (newOrbctlCommandFunc, error) {
	newCmd, err := chore.Orbctl(false, false)
	if err != nil {
		return nil, err
	}

	version := newCmd()
	version.Args = append(version.Args, "--version")

	outWriter, outWrite := logWriter(logger.Infof)
	defer outWrite()
	version.Stdout = outWriter

	errWriter, errWrite := logWriter(logger.Errorf)
	defer errWrite()
	version.Stderr = errWriter

	if err := version.Run(); err != nil {
		return nil, err
	}

	return func() (*exec.Cmd, error) {
		cmd := newCmd()
		cmd.Args = append(cmd.Args, "--orbconfig", orbconfig)
		return cmd, nil
	}, nil
}
