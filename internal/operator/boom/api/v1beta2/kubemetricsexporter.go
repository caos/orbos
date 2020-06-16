package v1beta2

type KubeMetricsExporter struct {
	Deploy       bool `json:"deploy" yaml:"deploy"`
	ReplicaCount int  `json:"replicaCount,omitempty" yaml:"replicaCount,omitempty"`
}
