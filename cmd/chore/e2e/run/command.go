package main

import (
	"bufio"
	"os/exec"
)

func simpleRunCommand(cmd *exec.Cmd, scan func(line string)) error {
	out, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		scan(scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return cmd.Wait()
}
