package v1beta2

import "github.com/caos/orbos/internal/operator/boom/api/v1beta2/toleration"

type MetricCollection struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy" yaml:"deploy"`
	//NodeSelector for deployment
	NodeSelector map[string]string `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	//Tolerations to run prometheus-operator on nodes
	Tolerations toleration.Tolerations `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`
}
