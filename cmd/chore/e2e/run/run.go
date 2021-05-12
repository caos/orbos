package main

import (
	"fmt"
	"io/ioutil"
	"reflect"
	"runtime"
	"strings"
	"time"

	"github.com/afiskon/promtail-client/promtail"
)

func runFunc(logger promtail.Client, branch, orbconfig string, from int, cleanup bool) func() error {
	return func() (err error) {

		newOrbctl, err := buildOrbctl(logger, orbconfig)
		if err != nil {
			return err
		}

		destroyTest := destroyTestFunc(logger)

		if cleanup {
			defer func() {
				if cuErr := destroyTest(newOrbctl, nil); cuErr != nil {

					original := ""
					if err != nil {
						original = fmt.Sprintf(": %s", err.Error())
					}

					err = fmt.Errorf("cleaning up after end-to-end test failed: %w%s", cuErr, original)
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

		readKubeconfig, deleteKubeconfig := readKubeconfigFunc(logger, kubeconfig.Name())
		defer deleteKubeconfig()

		branchParts := strings.Split(branch, "/")
		branch = branchParts[len(branchParts)-1:][0]

		return seq(logger, newOrbctl, configureKubectl(kubeconfig.Name()), from, readKubeconfig,
			/*  1 */ retry(3, initORBITERTest(logger, branch)),
			/*  2 */ retry(3, destroyTest),
			/*  3 */ retry(3, initBOOMTest(logger, branch)),
			/*  4 */ retry(3, bootstrapTestFunc(logger)),
			/*  5 */ ensureORBITERTest(logger, 5*time.Minute, isEnsured(1, 1, "v1.18.8")),
			/*  6 */ retry(3, patchTestFunc(logger, "clusters.k8s.spec.controlplane.nodes", "3")),
			/*  7 */ retry(3, patchTestFunc(logger, "clusters.k8s.spec.workers.0.nodes", "3")),
			/*  8 */ ensureORBITERTest(logger, 20*time.Minute, isEnsured(3, 3, "v1.18.8")),
			/*  9 */ retry(3, patchTestFunc(logger, "clusters.k8s.spec.versions.kubernetes", "v1.19.7")),
			/* 10 */ ensureORBITERTest(logger, 60*time.Minute, isEnsured(3, 3, "v1.19.7")),
			/* 11 */ retry(3, patchTestFunc(logger, "clusters.k8s.spec.controlplane.nodes", "1")),
			/* 12 */ retry(3, patchTestFunc(logger, "clusters.k8s.spec.workers.0.nodes", "1")),
			/* 13 */ ensureORBITERTest(logger, 60*time.Minute, isEnsured(1, 1, "v1.19.7")),
			/* 14 */ retry(3, patchTestFunc(logger, "clusters.k8s.spec.versions.kubernetes", "v1.20.2")),
			/* 15 */ ensureORBITERTest(logger, 60*time.Minute, isEnsured(1, 1, "v1.20.2")),
		)
	}
}

type testFunc func(newOrbctlCommandFunc, newKubectlCommandFunc) error

func seq(logger promtail.Client, orbctl newOrbctlCommandFunc, kubectl newKubectlCommandFunc, from int, readKubeconfigFunc func(orbctl newOrbctlCommandFunc) (err error), fns ...testFunc) error {

	var kcRead bool

	var at int
	for _, fn := range fns {
		at++
		if at < from {
			logger.Infof("\033[1;32mSkipping step %d\033[0m\n", at)
			continue
		}

		if at >= 5 && !kcRead {
			kcRead = true
			if err := readKubeconfigFunc(orbctl); err != nil {
				return err
			}
		}

		fnName := fmt.Sprintf("%s (%d. in stack)", runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name(), at)

		if err := fn(orbctl, kubectl); err != nil {
			return fmt.Errorf("%s failed: %w", fnName, err)
		}
		logger.Infof("\033[1;32m%s succeeded\033[0m\n", fnName)
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
