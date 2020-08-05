package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func run(orbconfig string) error {
	files, err := filepath.Glob("./cmd/chore/orbctl/*.go")

	if len(files) <= 0 {
		return errors.New("no files found in ./cmd/chore/orbctl/*.go")
	}

	if err != nil {
		panic(err)
	}

	args := []string{"run"}
	args = append(args, files...)
	args = append(args, "--orbconfig", orbconfig)
	args = append(args, "destroy")

	cmd := exec.Command("go", args...)
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		panic(err)
	}

	if err := cmd.Start(); err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := scanner.Text()
		fmt.Println(line)
		if strings.HasPrefix(line, "Are you absolutely sure") {
			if _, err := stdin.Write([]byte("y\n")); err != nil {
				panic(err)
			}
		}
	}
	if err := scanner.Err(); err != nil {
		panic(err)
	}
	if err := cmd.Wait(); err != nil {
		panic(err)
	}
	return nil
}
