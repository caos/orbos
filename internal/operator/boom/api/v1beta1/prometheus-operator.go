package v1beta1

import "github.com/caos/orbos/internal/operator/boom/api/v1beta1/toleration"

type PrometheusOperator struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy" yaml:"deploy"`
	//NodeSelector for deployment
	NodeSelector map[string]string `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	//Tolerations to run prometheus-operator on nodes
	Tolerations []toleration.Toleration `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`
}
