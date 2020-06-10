package v1beta2

type PostApply struct {
	Deploy bool   `json:"deploy,omitempty" yaml:"deploy,omitempty"`
	Folder string `json:"folder,omitempty" yaml:"folder,omitempty"`
}
