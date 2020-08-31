package v1beta2

import (
	"github.com/caos/orbos/internal/operator/boom/api/v1beta2/toleration"
	corev1 "k8s.io/api/core/v1"
)

type KubeMetricsExporter struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy" yaml:"deploy"`
	//Number of replicas used for deployment
	//@default: 1
	ReplicaCount int `json:"replicaCount,omitempty" yaml:"replicaCount,omitempty"`
	//NodeSelector for deployment
	NodeSelector map[string]string `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	//Tolerations to run kube state metrics exporter on nodes
	Tolerations toleration.Tolerations `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`
	//Resource requirements
	Resources *corev1.ResourceRequirements `json:"resources,omitempty" yaml:"resources,omitempty"`
}
