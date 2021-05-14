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

			file, err := os.Create(to)
			if err != nil {
				return err
			}
			defer file.Close()

			return runCommand(logger, orbctl(readsecretCtx), fmt.Sprintf("--gitops readsecret orbiter.%s.kubeconfig.encrypted", orb), false, file, nil)

		}, func() error {
			return os.Remove(to)
		}
}
