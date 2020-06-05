package v1beta1

type PostApply struct {
	//Flag if tool should be deployed
	Deploy bool `json:"deploy,omitempty" yaml:"deploy,omitempty"`
	//Relative path of folder in cloned git repository which should be applied
	Folder string `json:"folder,omitempty" yaml:"folder,omitempty"`
}
