package latest

import (
	"github.com/caos/orbos/internal/operator/boom/api/latest/storage"
	"github.com/caos/orbos/pkg/kubernetes/k8s"
)

type LogCollection struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy" yaml:"deploy"`
	//Fluentd Specs
	Fluentd *Fluentd `json:"fluentd,omitempty" yaml:"fluentd,omitempty"`
	//Fluentbit Specs
	Fluentbit *Component `json:"fluentbit,omitempty" yaml:"fluentbit,omitempty"`
	//Logging operator Specs
	Operator *Component `json:"operator,omitempty" yaml:"operator,omitempty"`
	//ClusterOutputs used by BOOM managed flows. BOOM managed Loki doesn't need to be specified
	ClusterOutputs []string `json:"clusterOutputs,omitempty" yaml:"clusterOutputs,omitempty"`
	//Outputs used by BOOM managed flows. BOOM managed Loki doesn't need to be specified
	Outputs []string `json:"outputs,omitempty" yaml:"outputs,omitempty"`
	//Watch these namespaces
	WatchNamespaces []string `json:"watchNamespaces,omitempty" yaml:"watchNamespaces,omitempty"`
	//Override used image version
	OverwriteVersion string `json:"overwriteVersion,omitempty" yaml:"overwriteVersion,omitempty"`
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
	//@default: 1
	Replicas *int `json:"replicas,omitempty" yaml:"replicas,omitempty"`
}
