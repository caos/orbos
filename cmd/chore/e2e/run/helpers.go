package main

import (
	"bufio"
	"errors"
	"os/exec"
	"time"
)

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
loop:
	for scanner.Scan() {
		select {
		case <-timer.C:
			return errors.New("timeout")
		default:
			if !scan(scanner.Text()) {
				cmd.Process.Kill()
				break loop
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return cmd.Wait()
}
