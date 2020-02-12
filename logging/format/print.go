package format

import (
	"fmt"
	"strings"
)

func CommitRecord(stringMap map[string]string) string {
	fields := toCommitFields(stringMap)
	stringFields := make([]string, len(fields))
	for idx, field := range fields {
		if field.key == "msg" {
			stringFields[idx] = fmt.Sprintf("EVENT: %s", field.value)
			continue
		}
		if field.key == "err" {
			stringFields[idx] = fmt.Sprintf("ERROR: %s", field.value)
			continue
		}
		stringFields[idx] = fmt.Sprintf("%s=%s", field.key, field.value)
	}
	return strings.Join(stringFields, ",")
}

func LogRecord(stringMap map[string]string) string {
	logLine := ""
	for _, field := range toLogFields(stringMap) {
		var color string
		switch field.key {
		case "msg":
			color = "1;35"
		case "err":
			color = "1;31"
		default:
			color = "0;33"
		}

		logLine = fmt.Sprintf("%s %s=\x1b[%sm\"%v\"\x1b[0m", logLine, field.key, color, field.value)
	}

	return strings.TrimSpace(logLine) + "\n"
}
