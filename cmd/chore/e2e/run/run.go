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
			/*  4 */ ensureORBITERTest(5*time.Minute, false),
			/*  5 */ patchTestFunc("clusters.k8s.spec.controlplane.nodes", "3"),
			/*  6 */ waitTest(15*time.Second),
			/*  7 */ ensureORBITERTest(20*time.Minute, true),
			/*  8 */ patchTestFunc("clusters.k8s.spec.versions.kubernetes", "v0.18.0"),
			/*  9 */ waitTest(15*time.Second),
			/* 10 */ ensureORBITERTest(60*time.Minute, false),
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

		if err := fn(orbctl, kubectl); err != nil {
			return fmt.Errorf("%s (%d. in stack) failed: %w", runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name(), at, err)
		}
	}
	return nil
}
