package v1beta2

import corev1 "k8s.io/api/core/v1"

type NodeMetricsExporter struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy" yaml:"deploy"`
	//Resource requirements
	Resources *corev1.ResourceRequirements `json:"resources,omitempty" yaml:"resources,omitempty"`
}
