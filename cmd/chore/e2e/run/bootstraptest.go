package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/caos/orbos/internal/helpers"
)

func bootstrap(orbctl newOrbctlCommandFunc) (err error) {

	cmd, err := orbctl()
	if err != nil {
		return err
	}

	cmd.Args = append(cmd.Args, "takeoff")
	cmd.Stderr = os.Stderr

	timer := time.NewTimer(20 * time.Minute)
	defer timer.Stop()
	return helpers.Concat(err, simpleRunCommand(cmd, func(line string) bool {
		select {
		case <-timer.C:
			err = errors.New("bootstrapping timed out after 20 minutes")
			return false
		default:
			fmt.Println(line)
			return true
		}
	}))
}
