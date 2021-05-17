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

func readKubeconfigFunc(settings programSettings, to string) (func(orbctl newOrbctlCommandFunc) (err error), func() error) {
	return func(orbctl newOrbctlCommandFunc) (err error) {

			readsecretCtx, readsecretCancel := context.WithTimeout(settings.ctx, 30*time.Second)
			defer readsecretCancel()

			file, err := os.Create(to)
			if err != nil {
				return err
			}
			defer file.Close()

			return runCommand(settings, orbctl(readsecretCtx), fmt.Sprintf("--gitops readsecret orbiter.%s.kubeconfig.encrypted", settings.orbID), false, file, nil)

		}, func() error {
			return os.Remove(to)
		}
}
