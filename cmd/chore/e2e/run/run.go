package main

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os/exec"
	"reflect"
	"runtime"
	"syscall"
	"time"

	"github.com/afiskon/promtail-client/promtail"
)

var _ runFunc = run

func run(settings programSettings) error {

	newOrbctl, err := buildOrbctl(settings)
	if err != nil {
		return err
	}

	if settings.cleanup {
		defer func() {
			// context is probably cancelled as program is terminating so we create a new one here
			destroySettings := settings
			destroySettings.ctx = context.Background()
			if cuErr := destroyTestFunc(destroySettings, newOrbctl, nil, 99); cuErr != nil {

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

	readKubeconfig, deleteKubeconfig := readKubeconfigFunc(settings, kubeconfig.Name())
	defer deleteKubeconfig()

	return seq(settings, newOrbctl, configureKubectl(kubeconfig.Name()), 4, readKubeconfig,
		/*  1 */ retry(3, writeInitialDesiredStateTest),
		/*  2 */ retry(3, destroyTestFunc),
		/*  3 */ retry(3, bootstrapTestFunc),
		/*  4 */ ensureORBITERTest(10*time.Minute, isEnsured(3, 3, "v1.18.8")),
		/*  5 */ retry(3, patchTestFunc(fmt.Sprintf("clusters.%s.spec.controlplane.nodes", settings.orbID), "1")),
		/*  6 */ retry(3, patchTestFunc(fmt.Sprintf("clusters.%s.spec.workers.0.nodes", settings.orbID), "2")),
		/*  7 */ ensureORBITERTest(5*time.Minute, isEnsured(1, 2, "v1.18.8")),
		/*  8 */ retry(3, patchTestFunc(fmt.Sprintf("clusters.%s.spec.versions.kubernetes", settings.orbID), "v1.21.0")),
		/*  9 */ ensureORBITERTest(45*time.Minute, isEnsured(1, 2, "v1.21.0")),
	)
}

var _ fmt.Stringer = (*programSettings)(nil)

type programSettings struct {
	ctx                          context.Context
	logger                       promtail.Client
	orbID, branch, orbconfig     string
	from                         uint8
	cleanup, debugOrbctlCommands bool
}

func (p *programSettings) String() string {
	return fmt.Sprintf(`orbconfig=%s
orb=%s
branch=%s
from=%d
cleanup=%t`,
		p.orbconfig,
		p.orbID,
		p.branch,
		p.from,
		p.cleanup,
	)
}

type testFunc func(programSettings, newOrbctlCommandFunc, newKubectlCommandFunc, uint8) error

func seq(settings programSettings, orbctl newOrbctlCommandFunc, kubectl newKubectlCommandFunc, readKubeconfigFrom uint8, readKubeconfigFunc func(orbctl newOrbctlCommandFunc) (err error), fns ...testFunc) error {

	var at uint8
	for _, fn := range fns {
		at++
		if at < settings.from {
			settings.logger.Infof("\033[1;32mSkipping step %d\033[0m\n", at)
			continue
		}

		if at >= readKubeconfigFrom {
			if err := readKubeconfigFunc(orbctl); err != nil {
				return err
			}
		}

		fnName := fmt.Sprintf("%s (%d. in stack)", runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name(), at)

		if err := fn(settings, orbctl, kubectl, at); err != nil {
			return fmt.Errorf("%s failed: %w", fnName, err)
		}
		settings.logger.Infof("\033[1;32m%s succeeded\033[0m\n", fnName)
	}
	return nil
}

func retry(count uint8, fn testFunc) testFunc {
	return func(settings programSettings, newOrbctl newOrbctlCommandFunc, newKubectl newKubectlCommandFunc, step uint8) error {
		return try(count, settings, step, newOrbctl, newKubectl, fn)
	}
}

func try(count uint8, settings programSettings, step uint8, newOrbctl newOrbctlCommandFunc, newKubectl newKubectlCommandFunc, fn testFunc) error {
	err := fn(settings, newOrbctl, newKubectl, step)
	exitErr := &exec.ExitError{}
	if err != nil &&
		// Don't retry if context timed out or was cancelled
		!errors.Is(err, context.Canceled) &&
		!errors.Is(err, context.DeadlineExceeded) &&
		// Don't retry if commands context timed out or was cancelled
		(!errors.As(err, &exitErr) ||
			!exitErr.Sys().(syscall.WaitStatus).Signaled()) &&
		count > 0 {
		return try(count-1, settings, step, newOrbctl, newKubectl, fn)
	}
	return err
}
