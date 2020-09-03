package k8s

import (
	corev1 "k8s.io/api/core/v1"
)

type Resources corev1.ResourceRequirements

func (r *Resources) MarshalYAML() (interface{}, error) {
	return MarshalYAML(r)
}

func (r *Resources) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return UnmarshalYAML(r, unmarshal)
}
