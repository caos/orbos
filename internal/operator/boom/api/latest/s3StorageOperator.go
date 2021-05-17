package latest

import "github.com/caos/orbos/pkg/kubernetes/k8s"

type S3StorageOperator struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy" yaml:"deploy"`
	//Domain name of the kubernetes cluster where the operator is running
	ClusterDomain string `json:"clusterDomain" yaml:"clusterDomain"`
	//Resource requirements
	Resources *k8s.Resources `json:"resources,omitempty" yaml:"resources,omitempty"`
	//Overwrite used image
	OverwriteImage string `json:"overwriteImage,omitempty" yaml:"overwriteImage,omitempty"`
	//Overwrite used image version
	OverwriteVersion string `json:"overwriteVersion,omitempty" yaml:"overwriteVersion,omitempty"`
}
