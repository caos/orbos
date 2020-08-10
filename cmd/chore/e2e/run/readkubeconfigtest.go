package main

import (
	"os"
)

func readKubeconfigTestFunc(to string) (func(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) (err error), func() error) {
	return func(orbctl newOrbctlCommandFunc, _ newKubectlCommandFunc) (err error) {

			readsecret, err := orbctl()
			if err != nil {
				return err
			}

			readsecret.Args = append(readsecret.Args, "readsecret", "orbiter.k8s.kubeconfig")
			readsecret.Stderr = os.Stderr

			file, err := os.Open(to)
			if err != nil {
				return err
			}
			defer file.Close()

			readsecret.Stdout = file

			return readsecret.Run()

		}, func() error {
			return os.Remove(to)
		}
}
