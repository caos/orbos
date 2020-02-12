package base

import (
	"github.com/caos/orbiter/logging"
)

type logger struct {
	fields  map[string]interface{}
	verbose bool
	onLog   func(bool, map[string]string)
}

func New() logging.Logger {
	return &logger{
		fields:  make(map[string]interface{}),
		verbose: false,
	}
}

func (l logger) AddSideEffect(onLog func(bool, map[string]string)) logging.Logger {
	original := l.onLog
	if original == nil {
		l.onLog = onLog
		return l
	}
	l.onLog = func(event bool, fields map[string]string) {
		original(event, fields)
		onLog(event, fields)
	}
	return l
}

func (l logger) IsVerbose() bool {
	return l.verbose
}

func (l logger) Verbose() logging.Logger {
	l.verbose = true
	return l
}

func (l logger) WithFields(fields map[string]interface{}) logging.Logger {
	l.fields = merge(l.fields, fields)
	return l
}

func merge(fields map[string]interface{}, add map[string]interface{}) map[string]interface{} {
	newFields := make(map[string]interface{})
	for k, v := range add {
		newFields[k] = v
	}
	for k, v := range fields {
		newFields[k] = v
	}
	return newFields
}

func (l logger) Info(event bool, msg string) {
	l.sideEffect(event, "msg", msg)
}

func (l logger) Debug(msg string) {
	if l.verbose {
		l.sideEffect(false, "debug", msg)
	}
}

func (l logger) Error(err error) {
	l.sideEffect(true, "err", err.Error())
}

func (l logger) sideEffect(event bool, key, value string) {
	if l.onLog != nil {
		l.onLog(event, toStringMap(merge(l.fields, map[string]interface{}{key: value})))
	}
}
