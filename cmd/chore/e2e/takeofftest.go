package main

import (
	"fmt"
	"os"
)

func takeoff(orbctl simpleOrbctlFunc) error {

	cmd, run, err := orbctl()
	if err != nil {
		return err
	}

	cmd.Args = append(cmd.Args, "takeoff")
	cmd.Stderr = os.Stderr
	return run(func(line string) bool {
		fmt.Println(line)
		return true
	})
}
