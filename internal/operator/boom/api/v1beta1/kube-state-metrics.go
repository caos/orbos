package v1beta1

type KubeStateMetrics struct {
	Deploy       bool `json:"deploy" yaml:"deploy"`
	ReplicaCount int  `json:"replicaCount,omitempty" yaml:"replicaCount,omitempty"`
}
