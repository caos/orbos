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
	k8sMap := make(map[string]interface{})
	return k8sMap, yaml.Unmarshal(lBytes, k8sMap)
}
