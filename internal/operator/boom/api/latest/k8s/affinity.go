package k8s

import (
	corev1 "k8s.io/api/core/v1"
)

type Affinity corev1.Affinity

func (a *Affinity) K8s() *corev1.Affinity {
	if a == nil {
		return nil
	}
	aff := corev1.Affinity(*a)
	return &aff
}

func (a *Affinity) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return UnmarshalYAML(a, unmarshal)
}

func (a *Affinity) MarshalYAML() (interface{}, error) {
	return MarshalYAML(a)
}
