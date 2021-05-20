package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

func runCommand(settings programSettings, log bool, write io.Writer, scan func(line string), cmd *exec.Cmd, args ...string) error {

	cmd.Args = append(cmd.Args, args...)

	errWriter, errWrite := logWriter(settings.logger.Warnf)
	defer errWrite()
	cmd.Stderr = errWriter

	stdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}

	var scanReader io.Reader = stdoutReader

	if write != nil {
		scanReader = io.TeeReader(stdoutReader, write)
	}

	settings.logger.Infof(fmt.Sprintf(`'%s'`, strings.Join(cmd.Args, `' '`)))

	if err := cmd.Start(); err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(scanReader)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		if scan != nil {
			scan(line)
		}

		if log {
			logORBITERStdout(settings, line)
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return cmd.Wait()
}
