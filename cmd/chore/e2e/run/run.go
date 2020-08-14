package main

import (
	"fmt"
	"io/ioutil"
	"strings"
	"time"
)

func runFunc(branch, orbconfig string, from int, cleanup bool) func() error {
	return func() error {

		newOrbctl, err := buildOrbctl(orbconfig)
		if err != nil {
			return err
		}

		if cleanup {
			defer func() {
				if cuErr := destroyTest(newOrbctl, nil); cuErr != nil {
					panic(cuErr)
				}
			}()
		}

		kubeconfig, err := ioutil.TempFile("", "kubeconfig-*")
		if err != nil {
			return err
		}
		if err := kubeconfig.Close(); err != nil {
			return err
		}

		readKubeconfig, deleteKubeconfig := readKubeconfigFunc(kubeconfig.Name())
		defer deleteKubeconfig()

		branchParts := strings.Split(branch, "/")
		branch = branchParts[len(branchParts)-1:][0]

		if err := seq(newOrbctl, configureKubectl(kubeconfig.Name()), from, readKubeconfig,
			/* 1 */ initORBITERTest(branch),
			/* 2 */ destroyTest,
			/* 3 */ bootstrapTest,
			/* 4 */ waitTest(15*time.Second),
			/* 5 */ ensureORBITERTest(5*time.Minute),
		); err != nil {
			return err
		}
		return nil
	}
}

func seq(orbctl newOrbctlCommandFunc, kubectl newKubectlCommandFunc, from int, readKubeconfigFunc func(orbctl newOrbctlCommandFunc) (err error), fns ...func(newOrbctlCommandFunc, newKubectlCommandFunc) error) error {

	var kcRead bool

	var at int
	for _, fn := range fns {
		at++
		if at < from {
			fmt.Println("Skipping step", at)
			continue
		}

		if at >= 6 && !kcRead {
			kcRead = true
			if err := readKubeconfigFunc(orbctl); err != nil {
				return err
			}
		}

		if err := fn(orbctl, kubectl); err != nil {
			return err
		}
	}
	return nil
}
