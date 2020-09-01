package toleration

import corev1 "k8s.io/api/core/v1"

type Tolerations []corev1.Toleration

/*
type Toleration struct {
	//Effect indicates the taint effect to match. Empty means match all taint effects. When specified, allowed values are NoSchedule, PreferNoSchedule and NoExecute.
	Effect string
	//Key is the taint key that the toleration applies to. Empty means match all taint keys. If the key is empty, operator must be Exists; this combination means to match all values and all keys.
	Key string
	//Operator represents a key's relationship to the value. Valid operators are Exists and Equal. Defaults to Equal. Exists is equivalent to wildcard for value, so that a pod can tolerate all taints of a particular category.
	Operator string
	//TolerationSeconds represents the period of time the toleration (which must be of effect NoExecute, otherwise this field is ignored) tolerates the taint. By default, it is not set, which means tolerate the taint forever (do not evict). Zero and negative values will be treated as 0 (evict immediately) by the system.
	TolerationSeconds *int64
	//Value is the taint value the toleration matches to. If the operator is Exists, the value should be empty, otherwise just a regular string.
	Value string
}

func (t *Toleration) ToKubeToleration() corev1.Toleration {
	return corev1.Toleration{
		Key:               t.Key,
		Operator:          corev1.TolerationOperator(t.Operator),
		Value:             t.Value,
		Effect:            corev1.TaintEffect(t.Effect),
		TolerationSeconds: t.TolerationSeconds,
	}
}

type Tolerations []*Toleration

func (t Tolerations) ToKubeToleartions() []corev1.Toleration {
	result := make([]corev1.Toleration, len(t), len(t))
	for idx, tol := range t {
		result[idx] = tol
	}
	return result
}
*/
