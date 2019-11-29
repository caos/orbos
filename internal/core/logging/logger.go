package logging

type Logger interface {
	Info(msg string)
	Debug(msg string)
	Error(err error)
	WithFields(map[string]interface{}) Logger
	Verbose() Logger
	IsVerbose() bool
}
