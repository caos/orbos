package kubebuilder

import (
	"github.com/caos/orbiter/logging"
	"github.com/go-logr/logr"
	"github.com/pkg/errors"
)

type logger struct {
	orbiter logging.Logger
}

func New(orbiterLogger logging.Logger) logr.Logger {
	return &logger{
		orbiter: orbiterLogger,
	}
}

func (l *logger) Error(err error, msg string, keysAndValues ...interface{}) {
	l.orbiter.WithFields(map[string]interface{}{}).Error(errors.Wrap(err, msg))
}

func (l *logger) V(level int) logr.InfoLogger {
	return l
}

func (l *logger) WithValues(keysAndValues ...interface{}) logr.Logger {
	return New(l.orbiter.WithFields(map[string]interface{}{}))
}

func (l *logger) WithName(name string) logr.Logger {
	return l
}

func (l *logger) Info(msg string, keysAndValues ...interface{}) {
	l.orbiter.WithFields(map[string]interface{}{}).Info(msg)
}

func (l *logger) Enabled() bool {
	return true
}
