package mntr

import "os"

func SprintCommit(_ string, fields map[string]string) string {
	return CommitRecord(AggregateCommitFields(fields))
}

func LogMessage(_ string, fields map[string]string) {
	log(fields)
}

func LogError(err error, fields map[string]string) {
	if err != nil {
		log(fields)
	}
}

func log(fields map[string]string) {
	WriteToStdout(LogRecord(AggregateLogFields(fields)))
}

func WriteToStdout(record string) {
	if _, err := os.Stdout.WriteString(record); err != nil {
		panic(err)
	}
}

func Concat(new func(string, map[string]string), original func(string, map[string]string)) func(string, map[string]string) {
	if original == nil {
		return new
	}

	return func(msg string, fields map[string]string) {
		new(msg, fields)
		original(msg, fields)
	}
}
