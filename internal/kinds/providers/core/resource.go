package core

import (
	"errors"
	"fmt"
)

type Resource struct {
	operatorID   *string
	service      ResourceService
	payload      interface{}
	desired      interface{}
	ensured      func(interface{}) error
	dependencies []*Resource
}

func newResource(operatorID *string, service ResourceService, payload interface{}, dependencies []*Resource, ensured func(interface{}) error) *Resource {
	return &Resource{operatorID, service, payload, nil, ensured, dependencies}
}

func (i *Resource) desire() (interface{}, error) {
	if i.service == nil {
		return nil, errors.New("No service specified for this resource")
	}
	var err error
	if i.desired == nil {
		i.desired, err = i.service.Desire(i.payload)
	}
	return i.desired, err
}

func (i *Resource) ID() (*string, error) {
	hashable, err := i.hashable()
	if err != nil {
		return nil, err
	}

	id := fmt.Sprintf("%s-%s-%s", *i.operatorID, i.service.Abbreviate(), hashable.hash())
	if len(id) > 63 {
		id = id[:63]
	}
	return &id, nil
}

func (i *Resource) hashable() (hashable, error) {

	hashableDeps := make([]hashable, 0)
	for _, depGroup := range i.dependencies {
		hashableDep, err := depGroup.hashable()
		if err != nil {
			return hashable{}, nil
		}
		hashableDeps = append(hashableDeps, hashableDep)
	}

	desired, err := i.desire()
	if err != nil {
		return hashable{}, err
	}

	return hashable{*i.operatorID, i.service.Abbreviate(), desired, hashableDeps}, nil
}
