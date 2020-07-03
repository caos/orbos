package v1beta2

type MetricCollection struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy" yaml:"deploy"`
}
