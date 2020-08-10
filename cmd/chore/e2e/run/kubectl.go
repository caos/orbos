package main

import (
	"os/exec"
)

type newKubectlCommandFunc func() *exec.Cmd

func configureKubectl(kubeconfig string) newKubectlCommandFunc {
	return func() *exec.Cmd {
		return exec.Command("kubectl", "--kubeconfig", kubeconfig)
	}
}
