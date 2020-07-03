package v1beta2

type KubeMetricsExporter struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy" yaml:"deploy"`
	//Number of replicas used for deployment
	//@default: 1
	ReplicaCount int `json:"replicaCount,omitempty" yaml:"replicaCount,omitempty"`
}
