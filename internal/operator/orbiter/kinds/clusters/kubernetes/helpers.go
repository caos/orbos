package kubernetes

import (
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/caos/orbiter/internal/helpers"
	"github.com/caos/orbiter/internal/operator/orbiter/kinds/clusters/core/infra"
	"github.com/caos/orbiter/mntr"
)

func try(monitor mntr.Monitor, timer *time.Timer, interval time.Duration, machine infra.Machine, callback func(cmp infra.Machine) error) error {
	var err error
	timedOut := helpers.Retry(timer, interval, func() bool {
		err = callback(machine)
		//		fmt.Println("SUCCESS")
		if err == nil {
			monitor.WithFields(map[string]interface{}{
				"machine": machine.ID(),
			}).Debug("retrying was successful")
			return false
		}
		//		 fmt.Printf("ERROR: %#v: %s\n", errors.Cause(cbErr), cbErr.Error())

		monitor := monitor.WithFields(map[string]interface{}{
			"cause": fmt.Sprintf("%#+v\n", err),
		})
		if exitErr, ok := errors.Cause(err).(*exec.ExitError); ok {
			monitor.WithFields(map[string]interface{}{
				"machine": machine.ID(),
			}).Debug("retrying failed severely")
			err = errors.Errorf("%s\n%s", exitErr.Error(), string(exitErr.Stderr))
			return false
		}
		monitor.WithFields(map[string]interface{}{
			"machine": machine.ID(),
		}).Debug("retrying failed, retrying...")

		return true
	})
	if timedOut != nil {
		return errors.Wrapf(err, "execution on node %s timed out after %s", machine.ID(), interval)
	}
	return nil
}

func operateConcurrently(machines []infra.Machine, cb func(infra.Machine) error) error {
	var wg sync.WaitGroup
	wg.Add(len(machines))
	syncronizer := helpers.NewSynchronizer(&wg)
	for _, machine := range machines {
		go func(cmp infra.Machine) {
			syncronizer.Done(errors.Wrapf(cb(cmp), "operating concurrently on machine %s failed", cmp.ID()))
		}(machine)
	}
	wg.Wait()

	if syncronizer.IsError() {
		return errors.Wrapf(syncronizer, "operating concurrently on machines %s", infra.Machines(machines))
	}

	return nil
}
