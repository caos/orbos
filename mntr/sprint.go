package mntr

import (
	"fmt"
	"github.com/caos/orbos/internal/ingestion"
	"github.com/golang/protobuf/ptypes"
	structpb "github.com/golang/protobuf/ptypes/struct"
	"strings"
)

func EventRecord(namespace, evt string, fields map[string]string) *ingestion.EventRequest {
	return &ingestion.EventRequest{
		CreationDate: ptypes.TimestampNow(),
		Data: &structpb.Struct{
			Fields: protoStruct(fields),
		},
		Type: strings.ReplaceAll(strings.ToLower(fmt.Sprintf("%s.%s", namespace, evt)), " ", "."),
	}
}

func protoStruct(fields map[string]string) map[string]*structpb.Value {
	pstruct := make(map[string]*structpb.Value)
	for key, value := range fields {
		if key == "ts" {
			continue
		}
		pstruct[key] = &structpb.Value{Kind: &structpb.Value_StringValue{StringValue: value}}
	}
	return pstruct
}

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
		case "msg":
			fallthrough
		case "evt":
			color = "1;35"
		case "err":
			color = "1;31"
		default:
			color = "0;33"
		}

		logLine = fmt.Sprintf("%s %s=\x1b[%sm\"%v\"\x1b[0m", logLine, field.Key, color, field.Value)
	}

	return strings.TrimSpace(logLine) + "\n"
}
