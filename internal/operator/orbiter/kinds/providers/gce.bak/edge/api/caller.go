package api

import (
	"context"
	"reflect"
)

type Caller struct {
	Ctx        context.Context
	OperatorID string
}

func (c *Caller) addContext(call reflect.Value) {
	call.MethodByName("Context").Call([]reflect.Value{reflect.ValueOf(c.Ctx)})
}
