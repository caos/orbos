package v1beta1

import (
	"github.com/caos/orbos/v5/internal/operator/boom/api/v1beta1/storage"
	"github.com/caos/orbos/v5/internal/operator/boom/api/v1beta1/toleration"
)

type LoggingOperator struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy" yaml:"deploy"`
	//Spec to define how the persistence should be handled
	FluentdPVC *storage.Spec `json:"fluentdStorage,omitempty" yaml:"fluentdStorage,omitempty"`
	//NodeSelector for deployment
	NodeSelector map[string]string `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	//Tolerations to run fluentbit on nodes
	Tolerations []toleration.Toleration `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`
}
