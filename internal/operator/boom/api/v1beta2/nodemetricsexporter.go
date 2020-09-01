package v1beta2

import "github.com/caos/orbos/internal/operator/boom/api/v1beta2/resources"

type NodeMetricsExporter struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy" yaml:"deploy"`
	//Resource requirements
	Resources *resources.Resources `json:"resources,omitempty" yaml:"resources,omitempty"`
}
