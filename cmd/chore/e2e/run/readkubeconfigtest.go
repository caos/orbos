package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/afiskon/promtail-client/promtail"
)

func readKubeconfigFunc(ctx context.Context, logger promtail.Client, orb, to string) (func(orbctl newOrbctlCommandFunc) (err error), func() error) {
	return func(orbctl newOrbctlCommandFunc) (err error) {

			readsecretCtx, readsecretCancel := context.WithTimeout(ctx, 30*time.Second)
			defer readsecretCancel()

			readsecret, err := orbctl(readsecretCtx)
			if err != nil {
				return err
			}

			readsecret.Args = append(readsecret.Args, "--gitops", "readsecret", fmt.Sprintf("orbiter.%s.kubeconfig.encrypted", orb))
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
