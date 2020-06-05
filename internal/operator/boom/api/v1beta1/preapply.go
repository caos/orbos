package v1beta1

type PreApply struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy,omitempty" yaml:"deploy,omitempty"`
	//Relative path of folder in cloned git repository which should be applied
	Folder string `json:"folder,omitempty" yaml:"folder,omitempty"`
}
