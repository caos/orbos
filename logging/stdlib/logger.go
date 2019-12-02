package stdlib

import (
	"io"

	"github.com/caos/orbiter/logging"
)

type logger struct {
	out     io.Writer
	fields  mapFields
	verbose bool
}

type WriterFunc func(p []byte) (n int, err error)

func (w WriterFunc) Write(p []byte) (n int, err error) {
	return w(p)
}

func New(out io.Writer) logging.Logger {
	return &logger{
		out:     out,
		fields:  make(map[string]interface{}),
		verbose: false,
	}
}
func (l logger) IsVerbose() bool {
	return l.verbose
}

func (l logger) Verbose() logging.Logger {
	l.verbose = true
	return l
}

func (l logger) WithFields(fields map[string]interface{}) logging.Logger {
	l.fields = l.fields.merge(fields)
	return l
}

func (l logger) Info(msg string) {
	l.print("msg", msg)
}

func (l logger) Debug(msg string) {
	if l.verbose {
		l.print("msg", msg)
	}
}

func (l logger) Error(err error) {
	l.print("err", err.Error())
}

func (l logger) print(key, value string) {
	fields := l.fields.merge(map[string]interface{}{
		key: value,
	})
	l.out.Write([]byte(fields.String()))
}
