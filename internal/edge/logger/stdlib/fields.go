package stdlib

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
)

type mapFields map[string]interface{}

func (f *mapFields) String() string {
	fields := toStructFields(*f)
	sort.Sort(ByPosition(fields))
	return marshal(fields)
}

func marshal(fields []field) string {

	line := ""
	for _, v := range fields {
		keyString := v.key
		rValue := reflect.ValueOf(v.value)
		rKind := rValue.Kind()
		switch rKind {
		case reflect.Map:
			stringKeyedMap := make(map[string]interface{})
			for _, rMapKey := range rValue.MapKeys() {
				stringKeyedMap[fmt.Sprintf("%s.%s", v.key, rMapKey.Interface())] = rValue.MapIndex(rMapKey).Interface()
			}
			line = fmt.Sprintf("%s %s", line, marshal(toStructFields(stringKeyedMap)))
		default:
			line = format(line, keyString, v.value)
		}
	}

	return strings.TrimSpace(line) + "\n"
}

func format(line string, key string, value interface{}) string {

	var color string
	switch key {
	case "msg":
		color = "1;35"
	case "err":
		color = "1;31"
	default:
		color = "0;33"
	}

	return fmt.Sprintf("%s %s=\x1b[%sm\"%v\"\x1b[0m", line, key, color, value)
}

func (f *mapFields) merge(fields map[string]interface{}) mapFields {
	newFields := make(map[string]interface{})
	for k, v := range fields {
		newFields[k] = v
	}
	for k, v := range *f {
		newFields[k] = v
	}
	return newFields
}
