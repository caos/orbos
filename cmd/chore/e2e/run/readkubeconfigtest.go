package main

import (
	"os"
)

func readKubeconfigFunc(to string) (func(orbctl newOrbctlCommandFunc) (err error), func() error) {
	return func(orbctl newOrbctlCommandFunc) (err error) {

			readsecret, err := orbctl()
			if err != nil {
				return err
			}

			readsecret.Args = append(readsecret.Args, "readsecret", "orbiter.k8s.kubeconfig")
			readsecret.Stderr = os.Stderr

			file, err := os.Create(to)
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
