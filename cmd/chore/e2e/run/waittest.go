package main

import (
	"time"
)

func waitTest(sleep time.Duration) func(newOrbctlCommandFunc, newKubectlCommandFunc) error {
	return func(_ newOrbctlCommandFunc, _ newKubectlCommandFunc) error {
		time.Sleep(sleep)
		return nil
	}
}
