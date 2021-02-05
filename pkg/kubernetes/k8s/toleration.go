package k8s

import corev1 "k8s.io/api/core/v1"

type Tolerations []Toleration

func (t Tolerations) K8s() []corev1.Toleration {
	if t == nil {
		return nil
	}
	tolerations := make([]corev1.Toleration, len(t))
	for idx := range t {
		tolerations[idx] = t[idx].K8s()
	}
	return tolerations
}

type Toleration corev1.Toleration

func (t Toleration) K8s() corev1.Toleration {
	return corev1.Toleration(t)
}

func (t Toleration) MarshalYAML() (interface{}, error) {
	return MarshalYAML(&t)
}

func (t *Toleration) UnmarshalYAML(unmarshal func(interface{}) error) error {
	return UnmarshalYAML(t, unmarshal)
}
