package latest

import (
	"github.com/caos/orbos/internal/operator/boom/api/latest/k8s"
	"github.com/caos/orbos/internal/operator/boom/api/latest/storage"
)

type LogCollection struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy" yaml:"deploy"`
	// Fluentd Specs
	Fluentd *Fluentd `json:"fluentd" yaml:"fluentd"`
	// Fluentbit Specs
	Fluentbit *Component `json:"fluentbit" yaml:"fluentbit"`
	// Logging operator Specs
	Operator *Component `json:"operator" yaml:"operator"`
}

type Component struct {
	//NodeSelector for deployment
	NodeSelector map[string]string `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	//Tolerations to run fluentbit on nodes
	Tolerations k8s.Tolerations `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`
	//Resource requirements
	Resources *k8s.Resources `json:"resources,omitempty" yaml:"resources,omitempty"`
}

type Fluentd struct {
	*Component `json:",inline" yaml:",inline"`
	//Spec to define how the persistence should be handled
	PVC *storage.Spec `json:"pvc,omitempty" yaml:"pvc,omitempty"`
	//Replicas number of fluentd instances
	Replicas int
}
