package v1beta1

type PreApply struct {
	Deploy bool   `json:"deploy,omitempty" yaml:"deploy,omitempty"`
	Folder string `json:"folder,omitempty" yaml:"folder,omitempty"`
}
