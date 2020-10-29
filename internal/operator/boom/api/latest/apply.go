package latest

//Apply: When the folder contains a kustomization.yaml-file the subfolders will be ignored. Otherwise all files inclusive the files contained by the subfolder will be applied if deploy=true, with deploy=false all will be deleted.
type Apply struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy,omitempty" yaml:"deploy,omitempty"`
	//Relative path of folder in cloned git repository which should be applied
	Folder string `json:"folder,omitempty" yaml:"folder,omitempty"`
}
