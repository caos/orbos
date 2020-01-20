package helpers

import (
	"fmt"
	"reflect"
)

func ToStringKeyedMap(m interface{}) (map[string]interface{}, error) {
	newMap := make(map[string]interface{})
	rValue := reflect.ValueOf(m)
	rKind := rValue.Kind()
	switch rKind {
	case reflect.Invalid:
		return newMap, nil
	case reflect.Map:
		for _, rMapKey := range rValue.MapKeys() {
			newMap[fmt.Sprintf("%s", rMapKey.Interface())] = rValue.MapIndex(rMapKey).Interface()
		}
	default:
		return nil, fmt.Errorf("Value %s is not a map", rKind)
	}
	return newMap, nil
}
