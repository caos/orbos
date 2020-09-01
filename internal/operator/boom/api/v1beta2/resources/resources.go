package resources

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
)

type Resources corev1.ResourceRequirements

func (r *Resources) MarshalYAML() (interface{}, error) {
	if r == nil {
		return nil, nil
	}

	intermediate, err := json.Marshal(r)
	if err != nil {
		return nil, err
	}

	result := make(map[string]interface{})

	return result, json.Unmarshal(intermediate, &result)
}
