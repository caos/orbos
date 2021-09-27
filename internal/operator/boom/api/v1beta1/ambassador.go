package v1beta1

import "github.com/caos/orbos/v5/internal/operator/boom/api/v1beta1/toleration"

type Ambassador struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy" yaml:"deploy"`
	//Number of replicas used for deployment
	//@default: 1
	ReplicaCount int `json:"replicaCount,omitempty" yaml:"replicaCount,omitempty"`
	//Service definition for ambassador
	Service *AmbassadorService `json:"service,omitempty" yaml:"service,omitempty"`
	//Activate the dev portal mapping
	ActivateDevPortal bool `json:"activateDevPortal,omitempty" yaml:"activateDevPortal,omitempty"`
	//NodeSelector for deployment
	NodeSelector map[string]string `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	//Tolerations to run ambassador on nodes
	Tolerations []toleration.Toleration `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`
}

type AmbassadorService struct {
	//Kubernetes service type
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	//IP when service is a loadbalancer with a fixed IP
	LoadBalancerIP string `json:"loadBalancerIP,omitempty" yaml:"loadBalancerIP,omitempty"`
	//Port definitions for the service
	Ports []*Port `json:"ports,omitempty" yaml:"ports,omitempty"`
}

type Port struct {
	//Name of the Port
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	//Port number
	Port uint16 `json:"port,omitempty" yaml:"port,omitempty"`
	//Targetport in-cluster
	TargetPort uint16 `json:"targetPort,omitempty" yaml:"targetPort,omitempty"`
	//Used port on node
	NodePort uint16 `json:"nodePort,omitempty" yaml:"nodePort,omitempty"`
}
