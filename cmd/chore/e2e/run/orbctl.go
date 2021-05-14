package main

import (
	"context"
	"os/exec"

	"github.com/afiskon/promtail-client/promtail"

	"github.com/caos/orbos/cmd/chore"
)

type newOrbctlCommandFunc func(context.Context) *exec.Cmd

func buildOrbctl(ctx context.Context, logger promtail.Client, orbconfig string) (newOrbctlCommandFunc, error) {
	newCmd, err := chore.Orbctl(false, false)
	if err != nil {
		return nil, err
	}

	if err := runCommand(logger, newCmd(ctx), "--version", true, nil, nil); err != nil {
		return nil, err
	}

	return func(ctx context.Context) *exec.Cmd {
		cmd := newCmd(ctx)
		cmd.Args = append(cmd.Args, "--orbconfig", orbconfig)
		return cmd
	}, nil
}
