package mntr

import (
	"fmt"
	"strings"
)

func CommitRecord(fields []*Field) string {
	stringFields := make([]string, len(fields))
	for _, field := range fields {
		if field.Key == "evt" {
			stringFields = append(stringFields, fmt.Sprint(field.Value))
			continue
		}
		if field.Key == "file" {
			continue
		}
		if field.Key == "err" {
			stringFields = append(stringFields, "An error occurred")
		}
		stringFields = append(stringFields, fmt.Sprintf("%s: %s", field.Key, field.Value))
	}
	return strings.Join(stringFields, "\n")
}

func LogRecord(fields []*Field) string {
	logLine := ""
	for _, field := range fields {
		var color string
		switch field.Key {
		case "msg", "evt":
			color = "1;35"
		case "err", "panic":
			color = "1;31"
		default:
			color = "0;33"
		}

		logLine = fmt.Sprintf("%s %s=\x1b[%sm\"%v\"\x1b[0m", logLine, field.Key, color, field.Value)
	}

	return strings.TrimSpace(logLine) + "\n"
}
