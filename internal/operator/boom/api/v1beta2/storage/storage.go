package storage

type Spec struct {
	//Defined size of the PVC
	StorageClass string `json:"storageClass,omitempty" yaml:"storageClass,omitempty"`
	//Storageclass used by the PVC
	AccessModes []string `json:"accessModes,omitempty" yaml:"accessModes,omitempty"`
	//Accessmodes used by the PVC
	Size string `json:"size,omitempty" yaml:"size,omitempty"`
}
