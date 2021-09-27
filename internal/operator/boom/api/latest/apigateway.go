package latest

import (
	"github.com/caos/orbos/v5/pkg/kubernetes/k8s"
	"github.com/caos/orbos/v5/pkg/secret"
)

type APIGateway struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy" yaml:"deploy"`
	//Number of replicas used for deployment
	//@default: 1
	ReplicaCount int `json:"replicaCount,omitempty" yaml:"replicaCount,omitempty"`
	//Pod scheduling constrains
	Affinity *k8s.Affinity `json:"affinity,omitempty" yaml:"affinity,omitempty"`
	//Service definition for ambassador
	Service *AmbassadorService `json:"service,omitempty" yaml:"service,omitempty"`
	//Activate the dev portal mapping
	ActivateDevPortal bool `json:"activateDevPortal,omitempty" yaml:"activateDevPortal,omitempty"`
	//NodeSelector for deployment
	NodeSelector map[string]string `json:"nodeSelector,omitempty" yaml:"nodeSelector,omitempty"`
	//Tolerations to run ambassador on nodes
	Tolerations k8s.Tolerations `json:"tolerations,omitempty" yaml:"tolerations,omitempty"`
	//Resource requirements
	Resources *k8s.Resources `json:"resources,omitempty" yaml:"resources,omitempty"`
	//Caching options
	Caching *Caching `json:"caching,omitempty" yaml:"caching,omitempty"`
	//Enable gRPC Web
	//@default: false
	GRPCWeb bool `json:"grpcWeb,omitempty" yaml:"grpcWeb,omitempty"`
	//Enable proxy protocol
	//@default: true
	ProxyProtocol *bool `json:"proxyProtocol,omitempty" yaml:"proxyProtocol,omitempty"`
	//Overwrite used image
	OverwriteImage string `json:"overwriteImage,omitempty" yaml:"overwriteImage,omitempty"`
	//Overwrite used image version
	OverwriteVersion string `json:"overwriteVersion,omitempty" yaml:"overwriteVersion,omitempty"`
	//License-key to use for Ambassador
	LicenceKey *secret.Secret `json:"licenceKey,omitempty" yaml:"licenceKey,omitempty"`
	//License-key to use for Ambassador
	ExistingLicenceKey *secret.Existing `json:"existingLicenceKey,omitempty" yaml:"existingLicenceKey,omitempty"`
}

func (a *APIGateway) IsZero() bool {
	if !a.Deploy &&
		a.ReplicaCount == 0 &&
		a.Affinity == nil &&
		a.Service == nil &&
		!a.ActivateDevPortal &&
		a.NodeSelector == nil &&
		a.Tolerations == nil &&
		a.Resources == nil &&
		a.Caching == nil &&
		!a.GRPCWeb &&
		a.ProxyProtocol == nil &&
		a.OverwriteVersion == "" &&
		(a.LicenceKey == nil || a.LicenceKey.IsZero()) &&
		a.ExistingLicenceKey == nil {
		return true
	}

	return false
}

func (a *APIGateway) InitSecrets() {
	if a.LicenceKey == nil {
		a.LicenceKey = &secret.Secret{}
		a.ExistingLicenceKey = &secret.Existing{}
	}
}

type Caching struct {
	//Enable specifies, whether a redis instance should be deployed or not
	Enable bool `json:"enable" yaml:"enable"`
	//Resource requirements
	Resources *k8s.Resources `json:"resources,omitempty" yaml:"resources,omitempty"`
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
