package main

import (
	"os"

	"github.com/afiskon/promtail-client/promtail"
)

func readKubeconfigFunc(logger promtail.Client, to string) (func(orbctl newOrbctlCommandFunc) (err error), func() error) {
	return func(orbctl newOrbctlCommandFunc) (err error) {

			readsecret, err := orbctl()
			if err != nil {
				return err
			}

			readsecret.Args = append(readsecret.Args, "--gitops", "readsecret", "orbiter.k8s.kubeconfig.encrypted")
			readsecretErrWriter, readsecretErrWrite := logWriter(logger.Errorf)
			defer readsecretErrWrite()
			readsecret.Stderr = readsecretErrWriter

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
