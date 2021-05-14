package main

import (
	"context"
	"os/exec"
)

type newKubectlCommandFunc func(context.Context) *exec.Cmd

func configureKubectl(kubeconfig string) newKubectlCommandFunc {
	return func(ctx context.Context) *exec.Cmd {
		return exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig)
	}
}
