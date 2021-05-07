package main

import (
	"bufio"
	"bytes"
	"io"
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
