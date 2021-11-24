package infra

import (
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/mntr"
)

func Try(monitor mntr.Monitor, timer *time.Timer, interval time.Duration, machine Machine, callback func(cmp Machine) error) error {
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
		var (
			exitErr   *exec.ExitError
			lookupErr *exec.Error
		)

		isExitErr := errors.As(err, &exitErr)
		isLookupErr := errors.As(err, &lookupErr)

		if isExitErr || isLookupErr {
			monitor.WithFields(map[string]interface{}{
				"machine": machine.ID(),
			}).Debug("retrying failed severely")
			if isExitErr {
				err = fmt.Errorf("%s: %w", string(exitErr.Stderr), exitErr)
			}
			return false
		}
		monitor.WithFields(map[string]interface{}{
			"machine": machine.ID(),
		}).Debug("retrying failed, retrying...")

		return true
	})

	if timedOut != nil {
		return fmt.Errorf("execution on node %s timed out after %s: %w", machine.ID(), interval, err)
	}
	return err
}
