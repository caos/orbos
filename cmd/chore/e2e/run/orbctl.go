package main

import (
	"context"
	"os/exec"

	"github.com/caos/orbos/cmd/chore/orbctl"
)

type newOrbctlCommandFunc func(context.Context) *exec.Cmd

func buildOrbctl(ctx context.Context, settings programSettings) (newOrbctlCommandFunc, error) {

	newCmd, err := orbctl.Command(false, false, settings.download, settings.tag)
	if err != nil {
		return nil, err
	}

	if err := runCommand(settings, orbctlPrefix.strPtr(), nil, nil, newCmd(ctx), "--version"); err != nil {
		return nil, err
	}

	return func(ctx context.Context) *exec.Cmd {
		cmd := newCmd(ctx)
		cmd.Args = append(cmd.Args, "--disable-analytics", "--orbconfig", settings.orbconfig)
		return cmd
	}, nil
}
