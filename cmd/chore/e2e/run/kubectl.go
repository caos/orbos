package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type newKubectlCommandFunc func(context.Context) *exec.Cmd

func configureKubectl(kubeconfig string) newKubectlCommandFunc {
	return func(ctx context.Context) *exec.Cmd {
		return exec.CommandContext(ctx, "kubectl", "--kubeconfig", kubeconfig)
	}
}

type downloadKubeconfig func(orbctl newOrbctlCommandFunc) (err error)

func downloadKubeconfigFunc(settings programSettings, to string) (downloadKubeconfig, func() error) {
	return func(orbctl newOrbctlCommandFunc) (err error) {

			readsecretCtx, readsecretCancel := context.WithTimeout(settings.ctx, 30*time.Second)
			defer readsecretCancel()

			file, err := os.Create(to)
			if err != nil {
				return err
			}
			defer file.Close()

			return runCommand(settings, false, file, nil, orbctl(readsecretCtx), "--gitops", "readsecret", fmt.Sprintf("orbiter.%s.kubeconfig.encrypted", settings.orbID))

		}, func() error {
			return os.Remove(to)
		}
}
