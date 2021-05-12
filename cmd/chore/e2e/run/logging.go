package main

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"github.com/afiskon/promtail-client/promtail"
)

func logWriter(log func(format string, args ...interface{})) (io.Writer, func()) {
	buf := new(bytes.Buffer)
	return buf, func() {
		defer buf.Reset()
		scanner := bufio.NewScanner(buf)
		for scanner.Scan() {
			log(scanner.Text())
		}
	}
}

func logORBITERStdout(logger promtail.Client, line string) {
	log := logger.Infof
	if strings.Contains(line, ` err=`) {
		log = logger.Warnf
	}
	log(line)
}
