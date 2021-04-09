package helpers

import "reflect"

func IsNil(sth interface{}) bool {
	return sth == nil || reflect.ValueOf(sth).Kind() == reflect.Ptr && reflect.ValueOf(sth).IsNil()
}
