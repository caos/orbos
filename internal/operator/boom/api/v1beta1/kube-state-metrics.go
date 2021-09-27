package v1beta1

import "github.com/caos/orbos/v5/internal/operator/boom/api/v1beta1/toleration"

type KubeStateMetrics struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy" yaml:"deploy"`
	//Number of replicas used for deployment
	//@default: 1
	ReplicaCount int `json:"replicaCount,omitempty" yaml:"replicaCount,omitempty"`
	//NodeSelector for deployment
	NodeSelector map[string]string `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	//Tolerations to run kube state metrics exporter on nodes
	Tolerations []toleration.Toleration `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`
}
