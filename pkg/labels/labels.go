package labels

import "gopkg.in/yaml.v3"

type Labels interface {
	comparable
	yaml.Marshaler
}

type comparable interface {
	Equal(comparable) bool
}

func K8sMap(l Labels) (map[string]interface{}, error) {
	lBytes, err := yaml.Marshal(l)
	if err != nil {
		return nil, err
	}
	mapOfStrings := make(map[string]string)
	k8sMap := make(map[string]interface{})
	if err := yaml.Unmarshal(lBytes, mapOfStrings); err != nil {
		return nil, err
	}
	for k, v := range mapOfStrings {
		k8sMap[k] = v
	}
	return k8sMap, nil
}

func MustK8sMap(l Labels) map[string]interface{} {
	m, err := K8sMap(l)
	if err != nil {
		panic(err)
	}
	return m
}
