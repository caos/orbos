package main

import (
	"bufio"
	"errors"
	"os/exec"
	"path/filepath"
)

type newOrbctlCommandFunc func() (*exec.Cmd, error)

func curryOrbctlCommand(orbconfig string) newOrbctlCommandFunc {
	return func() (*exec.Cmd, error) {
		return orbctlCommand(orbconfig)
	}
}

func orbctlCommand(orbconfig string) (*exec.Cmd, error) {
	files, _ := filepath.Glob("./cmd/chore/orbctl/*.go")
	if len(files) <= 0 {
		return nil, errors.New("no files found in ./cmd/chore/orbctl/*.go")
	}

	args := []string{"run"}
	args = append(args, files...)
	args = append(args, "--orbconfig", orbconfig)

	return exec.Command("go", args...), nil
}

func simpleRunCommand(cmd *exec.Cmd, scan func(line string) bool) error {
	out, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		if !scan(scanner.Text()) {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if err := cmd.Wait(); err != nil {
		return err
	}
	return nil
}
