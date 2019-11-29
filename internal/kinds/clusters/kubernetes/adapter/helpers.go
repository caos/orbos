package adapter

import (
	"fmt"
	"os/exec"
	"sync"
	"time"

	"github.com/pkg/errors"

	"github.com/caos/infrop/internal/core/helpers"
	"github.com/caos/infrop/internal/core/logging"
	"github.com/caos/infrop/internal/kinds/clusters/core/infra"
)

func try(logger logging.Logger, timer *time.Timer, interval time.Duration, compute infra.Compute, callback func(cmp infra.Compute) error) error {
	var err error
	timedOut := helpers.Retry(timer, interval, func() bool {
		err = callback(compute)
		//		fmt.Println("SUCCESS")
		if err == nil {
			logger.WithFields(map[string]interface{}{
				"compute": compute.ID(),
			}).Debug("retrying was successful")
			return false
		}
		//		 fmt.Printf("ERROR: %#v: %s\n", errors.Cause(cbErr), cbErr.Error())

		logger := logger.WithFields(map[string]interface{}{
			"cause": fmt.Sprintf("%#+v\n", err),
		})
		if exitErr, ok := errors.Cause(err).(*exec.ExitError); ok {
			logger.WithFields(map[string]interface{}{
				"compute": compute.ID(),
			}).Debug("retrying failed severely")
			err = errors.Errorf("%s\n%s", exitErr.Error(), string(exitErr.Stderr))
			return false
		}
		logger.WithFields(map[string]interface{}{
			"compute": compute.ID(),
		}).Debug("retrying failed, retrying...")

		return true
	})
	if timedOut != nil {
		return errors.Wrapf(err, "execution on node %s timed out after %s", compute.ID(), interval)
	}
	return nil
}

func operateConcurrently(computes []infra.Compute, cb func(infra.Compute) error) error {
	var wg sync.WaitGroup
	wg.Add(len(computes))
	syncronizer := helpers.NewSynchronizer(&wg)
	for _, compute := range computes {
		go func(cmp infra.Compute) {
			syncronizer.Done(errors.Wrapf(cb(cmp), "operating concurrently on compute %s failed", cmp.ID()))
		}(compute)
	}
	wg.Wait()

	if syncronizer.IsError() {
		return errors.Wrapf(syncronizer, "operating concurrently on computes %s", infra.Computes(computes))
	}

	return nil
}
