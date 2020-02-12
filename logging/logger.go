package logging

type Logger interface {
	Info(event bool, msg string)
	Debug(msg string)
	Error(err error)
	WithFields(map[string]interface{}) Logger
	Verbose() Logger
	IsVerbose() bool
	AddSideEffect(func(bool, map[string]string)) Logger
}
