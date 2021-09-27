package latest

import "github.com/caos/orbos/v5/pkg/kubernetes/k8s"

type NodeMetricsExporter struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy" yaml:"deploy"`
	//Resource requirements
	Resources *k8s.Resources `json:"resources,omitempty" yaml:"resources,omitempty"`
	//Overwrite used image
	OverwriteImage string `json:"overwriteImage,omitempty" yaml:"overwriteImage,omitempty"`
	//Overwrite used image version
	OverwriteVersion string `json:"overwriteVersion,omitempty" yaml:"overwriteVersion,omitempty"`
}
