package main

import (
	"bufio"
	"errors"
	"os/exec"
	"time"
)

var errTimeout = errors.New("timed out")

func simpleRunCommand(cmd *exec.Cmd, timer *time.Timer, scan func(line string) (goon bool)) error {
	defer timer.Stop()
	out, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		select {
		case <-timer.C:
			return errTimeout
		default:
			if !scan(scanner.Text()) {
				cmd.Process.Kill()
				return scanner.Err()
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return cmd.Wait()
}
