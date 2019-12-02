package context

import (
	"fmt"
	"runtime"
	"strings"
	"time"

	"github.com/caos/orbiter/logging"
)

type logger struct {
	verbose  bool
	original logging.Logger
}

func Add(l logging.Logger) logging.Logger {
	return &logger{l.IsVerbose(), l}
}

func (c logger) IsVerbose() bool {
	return c.verbose
}

func (c logger) Verbose() logging.Logger {
	c.original = c.original.Verbose()
	c.verbose = true
	return c
}

func (c logger) WithFields(fields map[string]interface{}) logging.Logger {
	c.original = c.original.WithFields(fields)
	return c
}

func (c logger) withTs() logging.Logger {
	return c.WithFields(map[string]interface{}{
		"ts": time.Now().Format(time.RFC3339),
	})
}

func (c logger) Error(err error) {
	l := c.withTs().(logger)
	if c.verbose {
		l = l.WithFields(callerFields()).(logger)
	}
	l.original.Error(err)
}

func (c logger) Info(msg string) {
	l := c.withTs().(logger)
	if c.verbose {
		l = l.WithFields(callerFields()).(logger)
	}
	l.original.Info(msg)
}

func (c logger) Debug(msg string) {
	if !c.verbose {
		// Take this responsibility for performance reasons
		return
	}
	c.withTs().WithFields(callerFields()).(logger).original.Debug(msg)
}

func callerFields() map[string]interface{} {

	pc := make([]uintptr, 15)
	n := runtime.Callers(1, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, more := frames.Next()

	seenCaller := false
framesLoop:
	for more && frame.Func != nil {
		receiverStart := strings.LastIndex(frame.Function, "/")
		receiverEnd := strings.LastIndex(frame.Function, ".")
		receiver := ""
		if receiverStart != -1 && receiverEnd != -1 {
			receiver = frame.Function[receiverStart+1 : receiverEnd]
		}

		if receiver != "context.logger" && receiver != "stdlib.logger" {
			if seenCaller {
				return map[string]interface{}{
					"file": fmt.Sprintf("%s:%d", frame.File, frame.Line),
				}
			}

			frame, more = frames.Next()
			continue framesLoop
		}
		seenCaller = true
		frame, more = frames.Next()
	}

	return make(map[string]interface{})
}
