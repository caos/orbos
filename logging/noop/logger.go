package noop

import "github.com/caos/orbiter/logging"

type logger struct{}

func New() logging.Logger { return &logger{} }

func (l *logger) AddSideEffect(func(bool, map[string]string)) logging.Logger { return l }
func (l *logger) WithFields(map[string]interface{}) logging.Logger           { return l }
func (l *logger) Info(bool, string)                                          {}
func (l *logger) Error(error)                                                {}
func (l *logger) Debug(string)                                               {}
func (l *logger) Verbose() logging.Logger                                    { return l }
func (l *logger) IsVerbose() bool                                            { return false }
