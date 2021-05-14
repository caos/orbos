package main

import (
	"bufio"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/afiskon/promtail-client/promtail"
)

func runCommand(logger promtail.Client, cmd *exec.Cmd, args string, log bool, write io.Writer, scan func(line string)) error {

	cmd.Args = append(cmd.Args, strings.Fields(args)...)

	errWriter, errWrite := logWriter(logger.Errorf)
	defer errWrite()
	cmd.Stderr = errWriter

	out, err := cmd.StdoutPipe()
	if err != nil {
		panic(err)
	}
	if err := cmd.Start(); err != nil {
		panic(err)
	}
	scanner := bufio.NewScanner(out)
	for scanner.Scan() {
		line := scanner.Text()
		if scan != nil {
			scan(line)
		}

		if log {
			logFunc := logger.Infof
			if strings.Contains(line, ` err=`) {
				logFunc = logger.Warnf
			}
			logFunc(line)
		}

		if write != nil {
			write.Write([]byte(fmt.Sprintf("%s\n", line)))
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return cmd.Wait()
}
