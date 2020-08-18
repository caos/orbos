package main

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"runtime"
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
			/*  1 */ initORBITERTest(branch),
			/*  2 */ destroyTest,
			/*  3 */ bootstrapTest,
			/*  4 */ ensureORBITERTest(5*time.Minute),
			/*  5 */ retry(3, patchTestFunc("clusters.k8s.spec.controlplane.nodes", "3")),
			/*  6 */ waitTest(15*time.Second),
			/*  7 */ retry(3, ensureORBITERTest(20*time.Minute)),
			/*  8 */ retry(3, patchTestFunc("clusters.k8s.spec.versions.kubernetes", "v0.18.0")),
			/*  9 */ waitTest(15*time.Second),
			/* 10 */ ensureORBITERTest(60*time.Minute),
			/* 11 */ ambassadorReadyTest,
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

		if at >= 4 && !kcRead {
			kcRead = true
			if err := readKubeconfigFunc(orbctl); err != nil {
				return err
			}
		}

		fnName := fmt.Sprintf("%s (%d. in stack)", runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name(), at)

		if err := fn(orbctl, kubectl); err != nil {
			return fmt.Errorf("%s failed: %w", fnName, err)
		}
		fmt.Printf("\033[1;32m%s succeeded\033[0m", fnName)
	}
	return nil
}

func retry(count uint8, fn func(newOrbctlCommandFunc, newKubectlCommandFunc) error) func(newOrbctlCommandFunc, newKubectlCommandFunc) error {
	return func(newOrbctl newOrbctlCommandFunc, newKubectl newKubectlCommandFunc) error {
		return try(count, newOrbctl, newKubectl, fn)
	}
}

func try(count uint8, newOrbctl newOrbctlCommandFunc, newKubectl newKubectlCommandFunc, fn func(newOrbctl newOrbctlCommandFunc, newKubectl newKubectlCommandFunc) error) error {
	err := fn(newOrbctl, newKubectl)
	if err != nil && count > 0 {
		return try(count-1, newOrbctl, newKubectl, fn)
	}
	return err
}
