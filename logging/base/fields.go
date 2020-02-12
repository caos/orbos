package base

import (
	"fmt"
	"reflect"
)

func toStringMap(fields map[string]interface{}) map[string]string {
	out := make(map[string]string)
	for key, value := range fields {
		rValue := reflect.ValueOf(value)
		rKind := rValue.Kind()
		switch rKind {
		case reflect.Map:
			stringKeyedMap := make(map[string]interface{})
			for _, rMapKey := range rValue.MapKeys() {
				stringKeyedMap[fmt.Sprintf("%s.%s", key, rMapKey.Interface())] = rValue.MapIndex(rMapKey).Interface()
			}
			for iKey, iValue := range toStringMap(stringKeyedMap) {
				out[iKey] = iValue
			}
		default:
			out[key] = fmt.Sprintf("%v", value)
		}
	}
	return out
}
