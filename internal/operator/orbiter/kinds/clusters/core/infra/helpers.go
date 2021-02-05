package infra

import (
	"fmt"
	"os/exec"
	"time"

	"github.com/caos/orbos/internal/helpers"
	"github.com/caos/orbos/mntr"
	"github.com/pkg/errors"
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
