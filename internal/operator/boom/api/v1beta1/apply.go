package v1beta1

type Apply struct {
	Deploy bool   `json:"deploy,omitempty" yaml:"deploy,omitempty"`
	Folder string `json:"folder,omitempty" yaml:"folder,omitempty"`
}
