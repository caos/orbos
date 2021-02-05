package labels

import (
	"gopkg.in/yaml.v3"
)

type Labels interface {
	comparable
	yaml.Marshaler
	yaml.Unmarshaler
	//	Major() int8
}

type IDLabels interface {
	Labels
	Name() string
}

type comparable interface {
	Equal(comparable) bool
}

func K8sMap(l Labels) (map[string]string, error) {
	return toMapOfStrings(l)
}

func MustK8sMap(l Labels) map[string]string {
	m, err := K8sMap(l)
	if err != nil {
		panic(err)
	}
	return m
}

func toMapOfStrings(sth interface{}) (map[string]string, error) {
	someBytes, err := yaml.Marshal(sth)
	if err != nil {
		return nil, err
	}
	mapOfStrings := make(map[string]string)
	return mapOfStrings, yaml.Unmarshal(someBytes, mapOfStrings)
}
