package api

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/caos/orbiter/internal/kinds/providers/core"
	"github.com/caos/orbiter/internal/core/helpers"

	"google.golang.org/api/googleapi"
)

func (c *Caller) GetResourceSelfLink(id string, getCalls []interface{}) (*string, error) {
	resource, err := c.GetResource(id, "selfLink", getCalls)
	if err != nil {
		return nil, err
	}

	if resource == nil {
		return nil, nil
	}

	selflink := reflect.Indirect(reflect.ValueOf(resource)).FieldByName("SelfLink").Interface().(string)
	return &selflink, nil
}

func (c *Caller) GetResource(id string, fields string, getCalls []interface{}) (interface{}, error) {

	var resource interface{}

	var wg sync.WaitGroup
	wg.Add(len(getCalls))
	synchronizer := helpers.NewSynchronizer(&wg)
	for _, getCall := range getCalls {
		go func(call interface{}) {
			value, err := c.callImmediately(call, fields, "")
			if err != nil {
				if googleErr, ok := err.(*googleapi.Error); ok && googleErr.Code == 404 {
					synchronizer.Done(nil)
					return
				}
			}
			if value != nil {
				resource = value.Interface()
			}
			synchronizer.Done(err)
		}(getCall)
	}
	wg.Wait()

	if synchronizer.IsError() {
		return nil, synchronizer
	}
	return resource, nil
}

func IsNotFound(err error) bool {

	if err == nil {
		return false
	}

	if googleErr, ok := err.(*googleapi.Error); ok && googleErr.Code == 404 {
		return true
	}
	return false
}

func (c *Caller) ListResources(resourceService core.ResourceService, listCalls []interface{}) ([]string, error) {
	names := make([]string, 0)

	var mux sync.RWMutex
	var wg sync.WaitGroup
	wg.Add(len(listCalls))
	synchronizer := helpers.NewSynchronizer(&wg)
	for _, listCall := range listCalls {
		go func(call interface{}) {
			ret, err := c.callImmediately(call, "items(name)", fmt.Sprintf("name:%s-%s-*", c.OperatorID, resourceService.Abbreviate()))
			if err == nil {
				listValue := reflect.Indirect(*ret).FieldByName("Items")
				for i := 0; i < listValue.Len(); i++ {
					name := reflect.Indirect(listValue.Index(i)).FieldByName("Name").Interface()
					mux.Lock()
					names = append(names, name.(string))
					mux.Unlock()
				}
			}
			synchronizer.Done(err)
		}(listCall)
	}
	wg.Wait()

	if synchronizer.IsError() {
		return nil, synchronizer
	}

	return names, nil
}

func (c *Caller) callImmediately(call interface{}, fields string, filter string) (*reflect.Value, error) {
	callValue := reflect.ValueOf(call)
	c.addContext(callValue)
	if fields != "" {
		callValue.MethodByName("Fields").Call([]reflect.Value{reflect.ValueOf(googleapi.Field(fields))})
	}
	if filter != "" {
		callValue.MethodByName("Filter").Call([]reflect.Value{reflect.ValueOf(filter)})
	}
	ret := callValue.MethodByName("Do").Call(nil)
	errValue := ret[1]
	if !errValue.IsNil() {
		return nil, errValue.Interface().(error)
	}
	return &ret[0], nil
}
