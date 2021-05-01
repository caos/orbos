package latest

type MetricsServer struct {
	//Flag if tool should be deployed
	//@default: false
	Deploy bool `json:"deploy" yaml:"deploy"`
	//Overwrite used image
	OverwriteImage string `json:"overwriteImage,omitempty" yaml:"overwriteImage,omitempty"`
	//Overwrite used image version
	OverwriteVersion string `json:"overwriteVersion,omitempty" yaml:"overwriteVersion,omitempty"`
}
