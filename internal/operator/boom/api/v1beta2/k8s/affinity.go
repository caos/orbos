package k8s

import (
	"encoding/json"

	corev1 "k8s.io/api/core/v1"
)

type Affinity corev1.Affinity

func (a *Affinity) UnmarshalYAML(unmarshal func(interface{}) error) error {
	generic := make(map[string]interface{})
	if err := unmarshal(&generic); err != nil {
		return err
	}

	intermediate, err := json.Marshal(generic)
	if err != nil {
		return err
	}

	return json.Unmarshal(intermediate, a)
}

func (a *Affinity) K8s() *corev1.Affinity {
	if a == nil {
		return nil
	}
	aff := corev1.Affinity(*a)
	return &aff
}
