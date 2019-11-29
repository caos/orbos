package noop

import "github.com/caos/orbiter/internal/core/logging"

type logger struct{}

func New() logging.Logger { return &logger{} }

func (l *logger) WithFields(map[string]interface{}) logging.Logger { return l }
func (l *logger) Info(msg string)                                  {}
func (l *logger) Debug(msg string)                                 {}
func (l *logger) Verbose() logging.Logger                          { return l }
func (l *logger) IsVerbose() bool                                  { return false }
