package v1beta1

type Ambassador struct {
	Deploy       bool               `json:"deploy" yaml:"deploy"`
	ReplicaCount int                `json:"replicaCount,omitempty" yaml:"replicaCount,omitempty"`
	Service      *AmbassadorService `json:"service,omitempty" yaml:"service,omitempty"`
}

type AmbassadorService struct {
	Type           string  `json:"type,omitempty" yaml:"type,omitempty"`
	LoadBalancerIP string  `json:"loadBalancerIP,omitempty" yaml:"loadBalancerIP,omitempty"`
	Ports          []*Port `json:"ports,omitempty" yaml:"ports,omitempty"`
}

type Port struct {
	Name       string `json:"name,omitempty" yaml:"name,omitempty"`
	Port       uint16 `json:"port,omitempty" yaml:"port,omitempty"`
	TargetPort uint16 `json:"targetPort,omitempty" yaml:"targetPort,omitempty"`
	NodePort   uint16 `json:"nodePort,omitempty" yaml:"nodePort,omitempty"`
}
