package v1beta2

import (
	"github.com/caos/orbos/internal/operator/boom/api/latest/k8s"
	"github.com/caos/orbos/internal/operator/boom/api/latest/storage"
)

type LogCollection struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy" yaml:"deploy"`
	//Spec to define how the persistence should be handled
	//@deprecated Use Fluentd.PVC instead
	FluentdPVC *storage.Spec `json:"fluentdStorage,omitempty" yaml:"fluentdStorage,omitempty"`
	//NodeSelector for deployment
	NodeSelector map[string]string `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	//Tolerations to run fluentbit on nodes
	Tolerations k8s.Tolerations `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`
	//Resource requirements
	Resources *k8s.Resources `json:"resources,omitempty" yaml:"resources,omitempty"`
}
